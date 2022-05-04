package internal

import (
	"fmt"
)

type Printer struct {
	configuration *Configuration
	storage       Storage
}

func NewPrinter(configuration *Configuration) *Printer {
	return &Printer{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (p *Printer) Print(messageId string) error {
	message, err := p.storage.Load(messageId)
	if err != nil {
		return err
	}

	if message != nil {
		fmt.Print(string(message.Raw))
	}

	return err
}
