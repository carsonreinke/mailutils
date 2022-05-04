package internal

type Storage interface {
	Save(message *Message) error
	Search(filter func(*Message) bool, ch chan<- *Message) error
	Load(messageId string) (*Message, error)
	Remove(messageId string) error
}
