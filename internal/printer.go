package internal

import (
	"fmt"
)

type Printer struct {
	configuration *Configuration
	storage Storage
}

func NewPrinter(configuration *Configuration) *Printer {
	return &Printer{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (p *Printer) Print(mailId string) error {
	message, err := p.storage.Load(mailId)
	if err != nil {
		return err
	}

	fmt.Print(string(message.Raw))

	return err
}