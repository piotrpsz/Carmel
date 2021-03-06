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

package responder

import (
	"Carmel/connector/datagram"
	"Carmel/connector/message"
	"Carmel/connector/tcpiface"
	"Carmel/secret/enigma"
	"Carmel/shared"
	"Carmel/shared/vtc"
	"log"
	"time"
)

type Responder struct {
	iface  *tcpiface.TCPInterface
	secret *enigma.Enigma
}

func New(iface *tcpiface.TCPInterface, e *enigma.Enigma) *Responder {
	return &Responder{iface: iface, secret: e}
}

func (r *Responder) Close() {
	if r.iface != nil {
		r.iface.Close()
		r.iface = nil
	}
}

func (r *Responder) IsValid(request, answer *message.Message, tstamp time.Time) bool {
	switch {
	case answer.Type != vtc.Answer:
		log.Printf("%s\n", "the message is not answer")
		return false
	case answer.Id != request.Id:
		log.Printf("%s\n", "invalid answer ID")
		return false
	case answer.Counter != request.Counter:
		log.Printf("%s\n", "invalid answer counter")
		return false
	case !(answer.Marker > request.Marker && shared.AreFloat32Equal(answer.Marker-3.1415, request.Marker)):
		log.Printf("%s\n", "invalid answer marker")
		return false
	case tstamp.Sub(answer.Tstamp).Seconds() > vtc.MessageTimeout:
		log.Printf("%s\n", "the answer was to long on the way")
		return false
	}
	return true
}

// Client side - read answer from the server
// The response must be the answer to the passed request
// ---------------------------------------------------
// 1. reading data from the network: message + signature
// 2. message and signature separation
// 3. signature verification
// 4. decrypting the message
// 5. unpacking JSON and converting it to a message object
// 6. checking the received message in terms of security
func (r *Responder) Read(request *message.Message) *message.Message {
	if data := datagram.Read(r.iface); data != nil { // 1.
		tstamp := shared.Now()
		if bytesCount := len(data); bytesCount > vtc.SignatureSize {
			sigIndex := bytesCount - vtc.SignatureSize
			cipher, sign := data[:sigIndex], data[sigIndex:] // 2.
			if r.secret.IsValidSignature(sign, cipher) {     // 3.
				if plain := r.secret.Decrypt(cipher); plain != nil { // 4.
					if answer := message.NewFromJson(plain); answer != nil { // 5.
						if r.IsValid(request, answer, tstamp) { // 6.
							return answer
						}
					}
				}
			}
		}
	}
	return nil
}

// Server side - send answer to the client.
// The response is for the passed request.
// ---------------------------------------------------
// 1. create a message
// 2. replace the message with JSON and pack (snapp)
// 3. data encryption
// 4. create a signature
// 5. adding a signature to an encrypted message
// 6. sending data to the network: message + signature
func (r *Responder) Send(status vtc.OperationStatusType, request *message.Message, data, extra []byte) *message.Message {
	if request == nil {
		log.Printf("%s\n", "invalid request")
		return nil
	}

	msg := message.NewWithType(vtc.Answer) // 1.
	msg.Counter = request.Counter
	msg.Marker = request.Marker + 3.1415
	msg.Id = request.Id
	msg.Data = data
	msg.Extra = extra
	msg.Tstamp = shared.Now()
	msg.Status = status

	if data := msg.ToJsonSnapped(); data != nil { // 2.
		if cipher := r.secret.Encrypt(data); cipher != nil { // 3.
			if sign := r.secret.Signature(cipher); sign != nil { // 4.
				cipher = append(cipher, sign...)    // 5.
				if datagram.Send(r.iface, cipher) { // 6.
					return msg
				}
			}
		}
	}
	return nil
}
