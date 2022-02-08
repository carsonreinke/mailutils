package internal

import (
	"path/filepath"
	"os"
	"io"
	"log"
	"errors"
	"strings"
	"bytes"
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

func hasMessageId(message *Message) bool {
	return message.Header != nil && strings.TrimSpace(message.Header.Get("Message-Id")) != ""
}

func (s *FileStorage) messageFilePath(message *Message) (string, error) {
	if !hasMessageId(message) {
		return "", errors.New("missing envelope and/or message id")
	}

	messageDate, err := message.Header.Date()
	if err != nil {
		return "", err
	}

	partition := filepath.Join(messageDate.Format("2006"), messageDate.Format("01")) 
	messageId := message.Header.Get("Message-Id")
	messageId = strings.ReplaceAll(messageId, "<", "")
	messageId = strings.ReplaceAll(messageId, ">", "")
	return filepath.Join(s.BasePath, partition, messageId + ".eml"), nil
}

func (s *FileStorage) load(filePath string) (*Message, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	message, err := NewMessage(file)
	return message, err
}

func (s *FileStorage) traverse(name string, current_depth int, ch chan<- *Message) error {
	entries, err := os.ReadDir(name)
	if err != nil {
		close(ch)
		return err
	}

	for _, entry := range entries {
		filePath := filepath.Join(name, entry.Name())
		if entry.IsDir() {
			if err := s.traverse(filePath, current_depth+1, ch); err != nil {
				close(ch)
				return err
			}
		} else {
			message, err := s.load(filePath)
			if err != nil {
				close(ch)
				return err
			}
			ch <- message
		}
	}

	if current_depth == 0 {
		close(ch)
	}
	return nil
}

func (s *FileStorage) Save(message *Message) error {
	if !hasMessageId(message) {
		subject := "(no header)"
		if message.Header != nil {
			subject = message.Header.Get("Subject")
		}
		log.Printf("Skipping message since it has no Message-Id: %s", subject)
		return nil
	}

	filePath, err := s.messageFilePath(message)
	if err != nil {
		return err
	}
	
	log.Printf("Saving message %s", message.Header.Get("Message-Id"))

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

	io.Copy(file, bytes.NewReader(message.Raw))

	return nil
}

func (s *FileStorage) Search(filter func(*Message) (bool), ch chan<- *Message) error {
	filter_ch := make(chan *Message)
	go func() {
		for message := range filter_ch {
			if filter(message) {
				ch <- message
			}
		}
		close(ch)
	}()
	
	err := s.traverse(s.BasePath, 0, filter_ch)
	if err != nil {
		close(filter_ch)
		close(ch)
	}
	return err
}

func (s *FileStorage) Load(mailId string) (*Message, error) {
	ch := make(chan *Message)
	done := make(chan error)
	go func() {
		err := s.traverse(s.BasePath, 0, ch)

		if err != nil {
			done <- err
		}
	}()
	
	for message := range ch {
		if(message.Header.Get("Message-Id") == mailId) {
			//close(ch)
			return message, nil
		}
	}
	if err := <-done; err != nil {
		return nil, err
	}
	return nil, nil
}

// func (s *FileStorage) Exists(mailId string) (*bool, error) {

// }
