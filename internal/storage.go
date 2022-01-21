package internal

type Storage interface {
	Save(message *Message) error
	Search(filter func(*Message) (bool), ch chan<- *Message) error
	Load(mailId string) (*Message, error)
	// Exists(mailId string) (*bool, error)
}
