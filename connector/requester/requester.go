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

package requester

import (
	"Carmel/connector/datagram"
	"Carmel/connector/message"
	"Carmel/connector/tcpiface"
	"Carmel/secret/enigma"
	"Carmel/shared"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"errors"
	"math/rand"
	"time"
)

type Requester struct {
	iface   *tcpiface.TCPInterface
	secret  *enigma.Enigma
	counter uint32
	marker  float32
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func New(iface *tcpiface.TCPInterface, e *enigma.Enigma) *Requester {
	return &Requester{iface: iface, secret: e}
}

func (r *Requester) IsValid(msg *message.Message, tstamp time.Time) bool {
	if msg.Type != vtc.Request {
		tr.IsOK(errors.New("message is not a request"))
		return false
	}
	if tstamp.Sub(msg.Tstamp).Seconds() > vtc.MessageTimeout {
		tr.IsOK(errors.New("the request was too long on the way"))
		return false
	}
	return true
}

// Client side - sending a request to the server.
// (The request is returned as a result)
// ----------------------------------------------
// 1. create a message object
// 2. Replace the message with JSON and compress it (snappy)
// 3. encryption
// 4. create a signature
// 5. sending data to the network: message + signature
func (r *Requester) Send(id uint32, data, extra []byte) *message.Message {
	r.counter++
	r.marker = float32(rand.Float64() * 99.9999)

	msg := message.NewWithType(vtc.Request) // 1.
	msg.Id = id
	msg.Counter = r.counter
	msg.Marker = r.marker
	msg.Tstamp = shared.Now()
	msg.Data = data
	msg.Extra = extra

	if data := msg.ToJsonSnapped(); data != nil { // 2.
		if cipher := r.secret.Encrypt(data); cipher != nil { // 3.
			if sign := r.secret.Signature(cipher); sign != nil { // 4.
				cipher = append(cipher, sign...) // 5.
				if datagram.Send(r.iface, cipher) {
					return msg
				}
			}
		}
	}
	return nil
}

// Server side - read request from client.
// (The request is returned as a result)
//----------------------------------------
// 1. reading data from the network: message + signature
// 2. message and signature separation
// 3. signature verification
// 4. data decryption
// 5. unpacking JSON and converting it to a message object
// 6. checking the received message in terms of security
func (r *Requester) Read() *message.Message {
	if data := datagram.Read(r.iface); data != nil { // 1.
		tstamp := shared.Now()
		if bytesCount := len(data); bytesCount > vtc.SignatureSize {
			sigIndex := bytesCount - vtc.SignatureSize
			cipher, sign := data[:sigIndex], data[sigIndex:] // 2.
			if r.secret.IsValidSignature(sign, cipher) {     // 3.
				if plain := r.secret.Decrypt(cipher); plain != nil { // 4.
					if msg := message.NewFromJson(plain); msg != nil { // 5.
						if r.IsValid(msg, tstamp) { // 6.
							return msg
						}
					}
				}
			}
		}
	}
	return nil
}
