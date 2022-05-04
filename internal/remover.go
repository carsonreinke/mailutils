package internal

import (
	"log"
)

type Remover struct {
	configuration *Configuration
	storage       Storage
}

func NewRemover(configuration *Configuration) *Remover {
	return &Remover{configuration: configuration, storage: NewFileStorage(configuration.StoragePath)}
}

func (p *Remover) Remove(messageIds []string) error {
	for _, messageId := range messageIds {
		if err := p.storage.Remove(messageId); err != nil {
			return err
		}

		log.Printf("Removed %v", messageId)
	}

	return nil
}
