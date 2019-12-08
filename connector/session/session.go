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
	"Carmel/secret/enigma"
	"Carmel/shared/tr"
	"encoding/json"
)

type Session struct {
	In     *stream.Stream
	Out    *stream.Stream
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
				tr.Info("Wysłano klucze")
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
							tr.Info("Odebrano i zainicjowano wszystkie klucze.")
							return true
						}
					}
				}
			}
		}
	}
	return false
}

/*
func (s *Session) ExchangeBlockIdentifiers(role vtc.RoleType) bool {
	switch role {
	case vtc.Server:
		return s.exchangeBlockIdentifiersAsServer()
	case vtc.Client:
		return s.exchangeBlockIdentifiersAsClient()
	}
	return false
}

func (s *Session) exchangeBlockIdentifiersAsServer() bool {
	// Serwer jako pierwszy wysyła swój blok identyfikacyjny
	if cipher := s.Enigma.EncryptRSA(s.Enigma.ServerId); cipher != nil {
		if !s.In.Requester.SendRawMessage(cipher) {
			return false
		}
	}

	// Serwer odczytuje blok identyfikacyjny klienta
	// i sprawdza jego poprawność.
	if cipher := s.; cipher != nil {
		if data := e.DecryptRsa(cipher); data != nil {
			if !secret.AreSlicesEqual(data, e.ClientId) {
				log.Println("invalid client identifier")
				return false
			}

		}
	}
	return true
}

func (e *Enigma) exchangeBlockIdentifiersAsClient(iface *tcpiface.TCPInterface) bool {
	// Klient odczytuje blok identyfikacyjny serwera
	// i sprawdza jego poprawność.
	if cipher := datagram.Read(iface); cipher != nil {
		if data := e.DecryptRsa(cipher); data != nil {
			if !secret.AreSlicesEqual(data, e.ServerId) {
				log.Println("invalid server identifier")
				return false
			}
		}
	}
	// Klient wysyła swój blok identyfikacyjny
	if cipher := e.EncryptRSA(e.ClientId); cipher != nil {
		if !datagram.Send(iface, cipher) {
			return false
		}
	}
	return true
}
*/
