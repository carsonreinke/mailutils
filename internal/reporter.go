package internal

import (
	"log"
	"sort"
)

type Reporter struct {
	configuration *Configuration
	storage       Storage
}

func NewReporter(configuration *Configuration) *Reporter {
	return &Reporter{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (r *Reporter) Report(header string) error {
	keys := make([]string, 0)
	report := make(map[string]uint)
	messages := make(chan *Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- r.storage.Search(func(message *Message) bool {
			return true
		}, messages)
	}()

	for message := range messages {
		value := message.Header.Get(header)
		if value == "" {
			continue
		}

		_, existing := report[value]
		if !existing {
			keys = append(keys, value)
			report[value] = 0
		}
		report[value] += 1
	}
	sort.Strings(keys)

	for _, key := range keys {
		log.Printf("%v: %v", key, report[key])
	}

	return <-done
}
