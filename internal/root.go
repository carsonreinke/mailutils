package internal

import (
	"time"
	"log"
)

type Configuration struct {
	Address string
	User string
	Password string
	IgnoreMailboxes []string `yaml:"ignore_mailboxes"`
	StoragePath string `yaml:"storage_path"`
	Timeout string
	TimeoutDuration time.Duration
}

func (c *Configuration) AssignDefaults() error {
	if c.Timeout == "" {
		c.Timeout = "60s"
	}
	var err error
	if c.TimeoutDuration, err = time.ParseDuration(c.Timeout); err != nil {
		return err
	}

	return nil
}

func Retry(attempts int, sleep time.Duration, function func() (interface{}, error)) (interface{}, error) {
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying after %v", err)
			time.Sleep(sleep * time.Duration(attempt))
		}

		var result interface{}
		result, err = function()
		if err == nil {
			return result, nil
		}
	}
	return nil, err
}