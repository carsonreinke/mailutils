package internal

import (
	"time"
)

type Configuration struct {
	Address string
	User string
	Password string
	IgnoreMailboxes []string `yaml:"ignore_mailboxes"`
	StoragePath string `yaml:"storage_path"`
}

func Retry(attempts int, sleep time.Duration, function func() error) error {
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			time.Sleep(sleep)
		}

		err = function()
		if err == nil {
			return nil
		}
	}
	return err
}