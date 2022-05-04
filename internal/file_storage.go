package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileStorage struct {
	BasePath string
}

func NewFileStorage(basePath string) *FileStorage {
	if basePath == "" {
		return nil
	}

	return &FileStorage{BasePath: basePath}
}

func hasMessageId(message *Message) bool {
	return message.Header != nil && strings.TrimSpace(message.Header.Get("Message-Id")) != ""
}

func messageFileBaseName(messageId string) string {
	messageId = strings.ReplaceAll(messageId, "<", "")
	messageId = strings.ReplaceAll(messageId, ">", "")
	return messageId + ".eml"
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
	return filepath.Join(s.BasePath, partition, messageFileBaseName(messageId)), nil
}

func (s *FileStorage) load(filePath string) (*Message, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	message, err := NewMessage(file)
	return message, err
}

func (s *FileStorage) find(name string, current_depth int, ch chan<- string, stop <-chan struct{}) error {
	entries, err := os.ReadDir(name)
	if err != nil {
		close(ch)
		return err
	}

	for _, entry := range entries {
		select {
		case <-stop:
			return nil
		default:
		}

		filePath := filepath.Join(name, entry.Name())
		if entry.IsDir() {
			if err := s.find(filePath, current_depth+1, ch, stop); err != nil {
				close(ch)
				return err
			}
		} else {
			ch <- filePath
		}
	}

	if current_depth == 0 {
		close(ch)
	}
	return nil
}

func (s *FileStorage) traverse(name string, ch chan<- *Message, stop <-chan struct{}) error {
	defer close(ch)

	var wg sync.WaitGroup
	concurrency := 5
	files_ch := make(chan string, concurrency)
	files_stop := make(chan struct{})
	done := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		done <- s.find(name, 0, files_ch, files_stop)
	}()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for filePath := range files_ch {
				select {
				case <-stop:
					return
				case <-files_stop:
					return
				default:
				}

				message, err := s.load(filePath)
				if err != nil {
					close(files_stop)
					done <- err
					return
				}
				ch <- message
			}
		}()
	}

	wg.Wait()
	return <-done
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

func (s *FileStorage) Search(filter func(*Message) bool, ch chan<- *Message) error {
	defer close(ch)

	filter_ch := make(chan *Message, 5)
	done := make(chan error, 1)

	go func() {
		done <- s.traverse(s.BasePath, filter_ch, make(chan struct{}))
	}()

	for message := range filter_ch {
		if filter(message) {
			ch <- message
		}
	}

	return <-done
}

func (s *FileStorage) Load(messageId string) (*Message, error) {
	ch := make(chan *Message)
	stop := make(chan struct{})
	done := make(chan error)

	go func() {
		done <- s.traverse(s.BasePath, ch, stop)
	}()

	for message := range ch {
		if message.Header.Get("Message-Id") == messageId {
			close(stop)
			return message, nil
		}
	}

	return nil, <-done
}

func (s *FileStorage) Remove(messageId string) error {
	ch := make(chan string)
	stop := make(chan struct{})
	done := make(chan error)

	go func() {
		done <- s.find(s.BasePath, 0, ch, stop)
	}()

	for filePath := range ch {
		//if filepath.Base(filePath) == messageFileBaseName(messageId) {
		if strings.HasSuffix(filePath, messageFileBaseName(messageId)) {
			err := os.Remove(filePath)
			return err
		}
	}

	if err := <-done; err != nil {
		return err
	}
	return fmt.Errorf("Message not found: %v", messageId)
}
