package message

import "strings"

type ChatMessage struct {
	Name string
	Text string
	Own  bool
}

func New(name, text string, own bool) ChatMessage {
	return ChatMessage{
		Name: strings.TrimSpace(name),
		Text: strings.TrimSpace(text),
		Own:  own,
	}
}

func (m ChatMessage) Valid() bool {
	return len(m.Text) > 0
}
