package internal

import (
	"net/mail"
)

type Storage interface {
	Save(message *Message) error
	Search(filter func(*mail.Message) (bool), ch chan<- *mail.Message) error
	// Load(mailId string) (*imap.Message, error)
	// Exists(mailId string) (*bool, error)
	// LoadAll(ch chan *imap.Message) error
}
