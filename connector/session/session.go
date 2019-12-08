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

package session

import (
	"Carmel/connector/stream"
	"Carmel/secret"
	"Carmel/secret/enigma"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"encoding/json"
)

type Session struct {
	In     *stream.Stream // klient -> serwer
	Out    *stream.Stream // serwer -> klient
	Enigma *enigma.Enigma
}

func ServerNew(port int) *Session {
	if e := enigma.New(""); e != nil {
		return &Session{In: stream.Server(port, e), Out: stream.Server(port+1, e), Enigma: e}
	}
	return nil
}

func ClientNew(addr string, port int, buddyName string, timeout int) *Session {
	if e := enigma.New(buddyName); e != nil {
		return &Session{In: stream.Client(addr, port+1, e, timeout), Out: stream.Client(addr, port, e, timeout), Enigma: e}
	}
	return nil
}

func (s *Session) Close() {
	defer func() {
		s.In = nil
		s.Out = nil
	}()
	s.In.Close()
	s.Out.Close()
	s.Enigma = nil
}

// Serwer wysyła do klienta wszystki klucze symetryczne.
func (s *Session) SendKeys() bool {
	defer s.Enigma.ClearKeys()

	if data, err := json.Marshal(s.Enigma.Keys); tr.IsOK(err) {
		if cipher := s.Enigma.EncryptRSA(data); cipher != nil {
			if s.Out.Requester.SendRawMessage(cipher) {
				return true
			}
		}
	}
	return false
}

// Klient odczytuje wszystkie klucze symetryczne od serwera.
func (s *Session) ReadKeys() bool {
	defer s.Enigma.ClearKeys()

	if cipher := s.In.Requester.ReadRawMessage(); cipher != nil {
		if data := s.Enigma.DecryptRsa(cipher); data != nil {
			if err := json.Unmarshal(data, &s.Enigma.Keys); tr.IsOK(err) {
				if s.Enigma.InitBlowfish(s.Enigma.Keys.Blowfish) {
					if s.Enigma.InitGhost(s.Enigma.Keys.Ghost) {
						if s.Enigma.InitWay3(s.Enigma.Keys.Way3) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (s *Session) ExchangeBlockIdentifiersAsServer() bool {
	// serwer żada blok identyfikujący od klienta (wysyła własny)
	if request := s.Out.Requester.Send(vtc.GetBlockID, s.Enigma.ServerId, nil); request != nil {
		if answer := s.Out.Responder.Read(request); answer != nil {
			if secret.AreSlicesEqual(answer.Data, s.Enigma.ClientId) {
				return true
			}
		}
	}
	return false
}

func (s *Session) ExchangeBlockIdentifiersAsClient() bool {
	// klient czeka na żądanie od serwera,
	// jeśli wszystko jest ok odsyła swój block identyfikujący
	if request := s.In.Requester.Read(); request != nil {
		if request.Id == vtc.GetBlockID {
			if secret.AreSlicesEqual(request.Data, s.Enigma.ServerId) {
				if answer := s.In.Responder.Send(vtc.Ok, request, s.Enigma.ClientId, nil); answer != nil {
					return true
				}
			}
		}
	}
	return false
}
