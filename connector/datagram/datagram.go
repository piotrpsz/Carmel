package datagram

import (
	"Carmel/shared/tr"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	Type    vtc.MessageType         `json:"type"`    // Request | Answer
	Id      uint32                  `json:"id"`      // identyfikator komunikatu (login, logout, ...)
	Status  vtc.OperationStatusType `json:"status"`  // status operacji, używane tylko w Answer
	Data    []byte                  `json:"data"`    // dane przesyłane w komunikacie
	Extra   []byte                  `json:"extra"`   // dodatkowe dane przesyłane w komunikacie
	Counter uint32                  `json:"counter"` // licznik komunikatów (kontrola bezpieczeństwa)
	Marker  float32                 `json:"marker"`  // marker (kontrola bezpieczeństwa)
	Tstamp  time.Time               `json:"tstamp"`  // stempel czasowy (kontrola bezpieczeństwa)
}

func NewWithType(msgType vtc.MessageType) *Message {
	if msgType == vtc.Request || msgType == vtc.Answer {
		return &Message{Type: msgType}
	}
	return nil
}

func NewFromJson(data []byte) *Message {
	return new(Message).fromSnappedJson(data)
}

func (msg *Message) ToJsonSnapped() []byte {
	if data, err := json.Marshal(msg); tr.IsOK(err) {
		return snappy.Encode(nil, data)
	}
	return nil
}

func (msg *Message) fromSnappedJson(data []byte) *Message {
	if data, err := snappy.Decode(nil, data); tr.IsOK(err) {
		if err := json.Unmarshal(data, msg); tr.IsOK(err) {
			if msg.Type == vtc.Request || msg.Type == vtc.Answer {
				return msg
			}
		}
	}
	return nil
}

func (msg *Message) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Message {\n")
	fmt.Fprintf(&b, "\t  Counter: %d\n", msg.Counter)
	fmt.Fprintf(&b, "\t       ID: %v\n", msg.Id)
	fmt.Fprintf(&b, "\t   Status: %v\n", msg.Status)
	fmt.Fprintf(&b, "\t     Type: %s\n", msg.typeAsString())
	fmt.Fprintf(&b, "\t     Data: %s\n", string(msg.Data))
	fmt.Fprintf(&b, "\t    Extra: %s\n", string(msg.Extra))
	fmt.Fprintf(&b, "\t   Marker: %v\n", msg.Marker)
	fmt.Fprintf(&b, "\tTimestamp: %v\n", shared.TimeAsString(msg.Tstamp))
	fmt.Fprintf(&b, "}")

	return b.String()
}

func (msg *Message) typeAsString() string {
	switch msg.Type {
	case vtc.Request:
		return "Request"
	case vtc.Answer:
		return "Answer"
	default:
		return "?"
	}
}
