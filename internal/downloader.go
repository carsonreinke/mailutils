package internal

import (
	"log"
    "math"
    "strings"
    "os"
    "fmt"
    "time"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

const retries = 5
const pageSize uint32 = 100
const limitDuration time.Duration = 100 * time.Millisecond

type Downloader struct {
	configuration *Configuration
	storage Storage
    debug bool
    client *client.Client
    limiter <-chan time.Time
}

func NewDownloader(configuration *Configuration) *Downloader {
    return &Downloader{
        configuration: configuration,
        storage: NewFileStorage(configuration.StoragePath),
        limiter: time.Tick(limitDuration),
        debug: false,
        client: nil,
    }
}

func (d *Downloader) Download() error {
    if err := d.assignClient(); err != nil {
        return err
    }

    i, err := d.retryWithConnection(func() (interface{}, error) {
        return d.getMailboxes()
    })
    if err != nil {
        return err
    }
    mailboxes := i.([]*imap.MailboxInfo)

    for _, mailbox := range mailboxes {
        ignore := false
        for _, ignoreMailbox := range d.configuration.IgnoreMailboxes {
            if strings.ToUpper(ignoreMailbox) == strings.ToUpper(mailbox.Name) {
                ignore = true
                break
            }
        }
        if ignore {
            log.Printf("Skipping mailbox %s", mailbox.Name)
            continue
        }

        log.Printf("Downloading mailbox %s", mailbox.Name)
        _, err := d.retryWithConnection(func() (interface{}, error) {
            err := d.downloadMailbox(mailbox)
            return nil, err
        })
        if err != nil {
            return err
        }
	}

    if d.client != nil {
        d.client.Logout()
        d.client = nil
    }
    
	return nil
}

func (d *Downloader) assignClient() error {
    var err error
    d.client, err = d.createClient()
    return err
}

func (d *Downloader) createClient() (*client.Client, error) {
    // TODO: Support non-TLS
    c, err := client.DialTLS(d.configuration.Address, nil)
    if err != nil {
		return nil, err
	}
    if d.debug {
        c.SetDebug(os.Stdout)
    }

	if err := c.Login(d.configuration.User, d.configuration.Password); err != nil {
		return c, err
	}

    if err := c.Noop(); err != nil {
		return c, err
	}

    if state := c.State(); state != 2 {
        return nil, fmt.Errorf("Connection is incorrect state of %d", state)
    }

    return c, nil
}

func (d *Downloader) useClient() (*client.Client, error) {
    if d.client == nil {
        if err := d.assignClient(); err != nil {
            return nil, err
        }
    }
    
    <-d.limiter
    return d.client, nil
}

func (d *Downloader) retryWithConnection(function func() (interface{}, error)) (interface{}, error) {
    return Retry(retries, d.configuration.TimeoutDuration, func() (interface{}, error) {
        result, err := function()

        // If there was an error, reset the connection
        if err != nil && d.client != nil {
            // Log out and ignore errors
            d.client.Logout()
            d.client = nil
        }

        return result, err
    })
}

func (d *Downloader) getMailboxes() ([]*imap.MailboxInfo, error) {
	mailboxes := make(chan *imap.MailboxInfo)
	done := make(chan error, 1)
	go func () {
        client, err := d.useClient()
        if err != nil {
            done <- err
        } else {
            done <- client.List("", "%", mailboxes)
        }
	}()

    mailboxesInfo := make([]*imap.MailboxInfo, 0)
	for mailbox := range mailboxes {
        mailboxesInfo = append(mailboxesInfo, mailbox)
    }
	if err := <-done; err != nil {
		return nil, err
	}

    return mailboxesInfo, nil
}

func (d *Downloader) downloadMailbox(info *imap.MailboxInfo) error {
    client, err := d.useClient()
    if err != nil {
        return err
    }
    mailbox, err := client.Select(info.Name, true)
    if err != nil {
        return err
    }
    log.Printf("There are %d messages to download", mailbox.Messages)

    // TODO: remove these crazy amount of casts
    pages := uint32(math.Ceil(float64(mailbox.Messages) / float64(pageSize)))

    from := uint32(1)
    for page := uint32(0); page < pages; page++ {
        to := from + pageSize - 1
        if to > mailbox.Messages {
            to = mailbox.Messages
        }

        _, err := d.retryWithConnection(func() (interface{}, error) {
            client, err := d.useClient()
            if err != nil {
                return nil, err
            }
            if _, err := client.Select(info.Name, true); err != nil {
                return nil, err
            }

            err = d.downloadRange(from, to)
            return nil, err
        })
        if err != nil {
            return err
        }

        from = to + 1
    }

    return nil
}

func (d *Downloader) downloadRange(from uint32, to uint32) error {
    log.Printf("Downloading message sequence from %d to %d", from, to)

    seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, pageSize)
	done := make(chan error, 1)
	go func() {
        client, err := d.useClient()
        if err != nil {
            done <- err
        } else {
		    done <- client.Fetch(seqset, []imap.FetchItem{imap.FetchRFC822}, messages)
        }
	}()
	
	for message := range messages {
        message, err := NewMessageFromIMAP(message)
        if err != nil {
            log.Printf("Skipping malformed message %v", message)
            continue
        }

		if err := d.storage.Save(message); err != nil {
		    if IsMalformedMessageError(err) {
                log.Printf("Skipping malformed message %v", message)
                continue
            }
            return err
		}
	}
	if err := <-done; err != nil {
		return err
	}

    return nil
}