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

func newMessageFromBuffer(buffer *bytes.Buffer) (*Message, error) {
	outbound := new(Message)
	outbound.Raw = buffer.Bytes()
	middle, err := mail.ReadMessage(bytes.NewReader(outbound.Raw))
	if err != nil {
		return nil, err
	}
	outbound.Message = middle

	return outbound, nil
}

func NewMessageFromIMAP(inbound *imap.Message) (*Message, error) {
	buffer := bytes.NewBuffer([]byte{})
	for _, value := range inbound.Body {
		io.Copy(buffer, value)
	}
	
	return newMessageFromBuffer(buffer)
}

func NewMessage(inbound io.Reader) (*Message, error) {
	buffer := bytes.NewBuffer([]byte{})
	io.Copy(buffer, inbound)
	
	return newMessageFromBuffer(buffer)
}

func IsMalformedMessageError(err error) bool {
	if err == mail.ErrHeaderNotPresent {
		return true
	}
	return false
}