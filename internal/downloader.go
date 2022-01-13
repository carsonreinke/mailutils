package internal

import (
	"log"
    "math"
    "strings"
    //"os"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
)

const retries = 3
const pageSize uint32 = 100

type Downloader struct {
	configuration *Configuration
	storage Storage
}

func NewDownloader(configuration *Configuration) *Downloader {
    return &Downloader{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (d *Downloader) Download() error {
    i, err := Retry(retries, d.configuration.TimeoutDuration, func() (interface{}, error) {
        return d.createClient()
    })
    if err != nil {
		return err
	}
    c := i.(*client.Client)
    if c != nil {
        defer c.Logout()
    }

	mailboxes := make(chan *imap.MailboxInfo, 5)
	done := make(chan error, 1)
	go func () {
		done <- c.List("", "%", mailboxes)
	}()

	for mailbox := range mailboxes {
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
		_, err := Retry(retries, d.configuration.TimeoutDuration, func() (interface{}, error) {
            err := d.downloadMailbox(c, mailbox)
            return nil, err
        })
        if err != nil {
            return err
        }
	}
	if err := <-done; err != nil {
		return err
	}

	return nil
}

func (d *Downloader) createClient() (*client.Client, error) {
    // TODO: Support non-TLS
    c, err := client.DialTLS(d.configuration.Address, nil)
    if err != nil {
		return nil, err
	}
    //c.SetDebug(os.Stdout)

	if err := c.Login(d.configuration.User, d.configuration.Password); err != nil {
		return c, err
	}

    if err := c.Noop(); err != nil {
		return c, err
	}

    return c, nil
}

func (d *Downloader) downloadMailbox(c *client.Client, info *imap.MailboxInfo) error {
    mailbox, err := c.Select(info.Name, true)
    if err != nil {
        return err
    }

    // TODO: remove these crazy amount of casts
    pages := uint32(math.Ceil(float64(mailbox.Messages) / float64(pageSize)))

    from := uint32(1)
    for page := uint32(0); page < pages; page++ {
        to := from + pageSize - 1
        if to > mailbox.Messages {
            to = mailbox.Messages
        }

        _, err := Retry(retries, d.configuration.TimeoutDuration, func() (interface{}, error) {
            err := d.downloadRange(c, from, to)
            return nil, err
        })
        if err != nil {
            return err
        }

        from = to + 1
    }

    return nil
}

func (d *Downloader) downloadRange(c *client.Client, from uint32, to uint32) error {
    log.Printf("Downloading message sequence from %d to %d", from, to)

    seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, pageSize)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchRFC822, imap.FetchEnvelope}, messages)
	}()
	
	for message := range messages {
		if err := d.storage.Save(message); err != nil {
            close(done)
			return err
		}
	}
	if err := <-done; err != nil {
		return err
	}

    return nil
}