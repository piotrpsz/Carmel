/*
 * BSD 2-Clause License
 *
 *	Copyright (c) 2019, Piotr Pszczółkowski
 *	All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

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
	Type    vtc.MessageType         `json:"type"`            // Request | Answer
	Id      uint32                  `json:"id"`              // message ID (login, logout, ...)
	Status  vtc.OperationStatusType `json:"status"`          // operation status, used only in answer
	Data    []byte                  `json:"data"`            // data sent in the message
	Extra   []byte                  `json:"extra,omitempty"` // additional data sent in the message
	Blob    []byte                  `json:"blob,omitempty"`  // string of bytes with context dependent meaning
	Counter uint32                  `json:"counter"`         // message counter (security check)
	Marker  float32                 `json:"marker"`          // marker (security check)
	Tstamp  time.Time               `json:"tstamp"`          // time stamp (security check)
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
