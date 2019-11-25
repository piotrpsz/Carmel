package message

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"encoding/json"
	"fmt"
	"github.com/golang/snappy"
	"strings"
	"time"
)

type Message struct {
	Type    vtc.MessageType         `json:"type"`    // Request | Answer
	Id      uint32                  `json:"id"`      // message ID (login, logout, ...)
	Status  vtc.OperationStatusType `json:"status"`  // operation status, used only in answer
	Data    []byte                  `json:"data"`    // data sent in the message
	Extra   []byte                  `json:"extra"`   // additional data sent in the message
	Counter uint32                  `json:"counter"` // message counter (security check)
	Marker  float32                 `json:"marker"`  // marker (security check)
	Tstamp  time.Time               `json:"tstamp"`  // time stamp (security check)
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

func (m *Message) ToJsonSnapped() []byte {
	if data, err := json.Marshal(m); tr.IsOK(err) {
		return snappy.Encode(nil, data)
	}
	return nil
}

func (m *Message) fromSnappedJson(data []byte) *Message {
	if data, err := snappy.Decode(nil, data); tr.IsOK(err) {
		if err := json.Unmarshal(data, m); tr.IsOK(err) {
			if m.Type == vtc.Request || m.Type == vtc.Answer {
				return m
			}
		}
	}
	return nil
}

func (m *Message) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "Message {\n")
	fmt.Fprintf(&b, "\t  Counter: %d\n", m.Counter)
	fmt.Fprintf(&b, "\t       ID: %v\n", m.Id)
	fmt.Fprintf(&b, "\t   Status: %v\n", m.Status)
	fmt.Fprintf(&b, "\t     Type: %s\n", m.typeAsString())
	fmt.Fprintf(&b, "\t     Data: %s\n", string(m.Data))
	fmt.Fprintf(&b, "\t    Extra: %s\n", string(m.Extra))
	fmt.Fprintf(&b, "\t   Marker: %v\n", m.Marker)
	fmt.Fprintf(&b, "\tTimestamp: %v\n", shared.TimeAsString(m.Tstamp))
	fmt.Fprintf(&b, "}")

	return b.String()
}

func (m *Message) typeAsString() string {
	switch m.Type {
	case vtc.Request:
		return "Request"
	case vtc.Answer:
		return "Answer"
	default:
		return "?"
	}
}
