package internal

import (
	"path/filepath"
	"os"
	"io"
	"log"
	"errors"
	"strings"
	//"bufio"

	"github.com/emersion/go-imap"
)

type FileStorage struct {
	BasePath string
}

func NewFileStorage(basePath string) (*FileStorage) {
	if basePath == "" {
		return nil
	}

	return &FileStorage{BasePath: basePath}
}

func (s *FileStorage) messageFilePath(message *imap.Message) (string, error) {
	if message.Envelope == nil || strings.TrimSpace(message.Envelope.MessageId) == "" {
		return "", errors.New("missing envelope and/or message id")
	}

	partition := filepath.Join(message.Envelope.Date.Format("2006"), message.Envelope.Date.Format("01")) 
	messageId := message.Envelope.MessageId
	messageId = strings.ReplaceAll(messageId, "<", "")
	messageId = strings.ReplaceAll(messageId, ">", "")
	return filepath.Join(s.BasePath, partition, messageId + ".eml"), nil
}

func (s *FileStorage) Save(message *imap.Message) error {
	filePath, err := s.messageFilePath(message)
	if err != nil {
		return err
	}
	
	log.Printf("Saving message %s", message.Envelope.MessageId)

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Sync()
	defer file.Close()

	//fileWriter := bufio.NewWriter(file)
	for _, value := range message.Body {
		io.Copy(file, value)
	}
	//fileWriter.Flush()

	return nil
}

// func (s *FileStorage) Load(mailId string) (*imap.Message, error) {

// }

// func (s *FileStorage) Exists(mailId string) (*bool, error) {

// }

// func (s *FileStorage) LoadAll(ch chan *imap.Message) error {

// }

