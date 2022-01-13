package internal

import (
	"github.com/emersion/go-imap"
)

type Storage interface {
	Save(message *imap.Message) error
	// Load(mailId string) (*imap.Message, error)
	// Exists(mailId string) (*bool, error)
	// LoadAll(ch chan *imap.Message) error
}
