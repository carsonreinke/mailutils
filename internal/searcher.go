package internal

import (
	"regexp"
	"net/mail"
	"log"
	"bufio"
)

type Searcher struct {
	configuration *Configuration
	storage Storage
}

func NewSearcher(configuration *Configuration) *Searcher {
	return &Searcher{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (d *Searcher) Search(keywords []string) error {
	regexps := make([]*regexp.Regexp, len(keywords))
	for index, keyword := range keywords {
		regexp, err := regexp.Compile(regexp.QuoteMeta(keyword))
		if err != nil {
			return err
		}
		regexps[index] = regexp
	}
	

	messages := make(chan *mail.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- d.storage.Search(func(message *mail.Message) bool {
			match := true
			for _, regexp := range regexps {
				match = match && regexp.MatchReader(bufio.NewReader(message.Body))
			}
			return match
		}, messages)
	}()

	for message := range messages {
		log.Printf("%v", message.Header.Get("Message-Id"))
	}
	if err := <-done; err != nil {
		return err
	}

	return nil
}