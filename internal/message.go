package internal

import (
	"net/mail"
	"io"
	"bytes"

	"github.com/emersion/go-imap"
)

type Message struct {
	*mail.Message

	Raw []byte
}

func NewMessage(inbound *imap.Message) (*Message, error) {
	buffer := bytes.NewBuffer([]byte{})
	for _, value := range inbound.Body {
		io.Copy(buffer, value)
	}
	
	outbound := new(Message)
	outbound.Raw = buffer.Bytes()
	middle, err := mail.ReadMessage(bytes.NewReader(outbound.Raw))
	if err != nil {
		return nil, err
	}
	outbound.Message = middle

	return outbound, nil
}