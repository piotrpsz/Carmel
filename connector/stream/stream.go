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

package stream

import (
	"Carmel/connector/requester"
	"Carmel/connector/responder"
	"Carmel/connector/tcpiface"
	"Carmel/secret"
	"Carmel/secret/enigma"
	"Carmel/secret/enigma/blowfish"
	"Carmel/secret/enigma/ghost"
	"Carmel/secret/enigma/way3"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Stream struct {
	role       vtc.RoleType
	Enigma     *enigma.Enigma
	Responder  *responder.Responder
	Requester  *requester.Requester
	RemoteAddr string
	ServerAddr string // client only
	ServerPort int    // server & client
	timeout    int    // client only
}

func Server(port int, e *enigma.Enigma) *Stream {
	s := &Stream{role: vtc.Server, ServerPort: port, Enigma: e}
	if s.InitKeys() {
		return s
	}
	return nil
}

func (s *Stream) InitKeys() bool {
	if key := secret.RandomBytes(blowfish.MaxKeyLength); s.Enigma.InitBlowfish(key) {
		if key := secret.RandomBytes(ghost.KeySize); s.Enigma.InitGhost(key) {
			if key := secret.RandomBytes(way3.KeySize); s.Enigma.InitWay3(key) {
				return true
			}
		}
	}
	return false
}

func Client(addr string, port int, e *enigma.Enigma, timeout int) *Stream {
	return &Stream{role: vtc.Client, ServerAddr: addr, ServerPort: port, Enigma: e, timeout: timeout}
}

func (s *Stream) Close() {
	defer func() {
		s.Responder = nil
		s.Requester = nil
	}()
	s.Responder.Close()
	s.Requester.Close()
}

/********************************************************************
*                                                                   *
*                             R U N                                 *
*                                                                   *
********************************************************************/

func (s *Stream) Run(ctx context.Context, wg *sync.WaitGroup) vtc.OperationStatusType {
	defer wg.Done()

	retChan := make(chan vtc.OperationStatusType)
	defer close(retChan)
	var iwg sync.WaitGroup

	switch s.role {
	//------- Server -------
	case vtc.Server:
		ictx, cancel := context.WithCancel(context.Background())
		iwg.Add(1)
		go s.runServer(ictx, &iwg, retChan)

		select {
		case <-ctx.Done():
			log.Println("Stream.Run.Server:", ctx.Err())
			cancel()
			iwg.Wait()
			return vtc.Cancel
		case retv := <-retChan:
			return retv
		}

	//------- Client -------
	case vtc.Client:
		ictx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
		iwg.Add(1)
		go s.runClient(ictx, &iwg, retChan)

		select {
		case <-ctx.Done():
			log.Println("Stream.Run.Client:", ctx.Err())
			cancel()
			iwg.Wait()
			return vtc.Cancel
		case retv := <-retChan:
			iwg.Wait()
			return retv
		}

	}
	return vtc.Error
}

func (s *Stream) runServer(ctx context.Context, wg *sync.WaitGroup, retChan chan<- vtc.OperationStatusType) {
	defer wg.Done()

	rc := make(chan vtc.OperationStatusType)
	go s.waitForClient(rc)

	select {
	case <-ctx.Done():
		log.Println("Stream.runServer:", ctx.Err())
	case retv := <-rc:
		retChan <- retv
	}
}

func (s *Stream) waitForClient(retChan chan<- vtc.OperationStatusType) {
	if hostname, err := os.Hostname(); tr.IsOK(err) {
		if addr, err := net.ResolveIPAddr("ip", hostname); tr.IsOK(err) {
			tcpAdrr := net.TCPAddr{IP: addr.IP, Port: s.ServerPort, Zone: addr.Zone}
			fmt.Printf("Server: %s:%d\n", tcpAdrr.IP, tcpAdrr.Port)

			if listener, err := net.ListenTCP("tcp", &tcpAdrr); tr.IsOK(err) {
				if conn, err := listener.AcceptTCP(); tr.IsOK(err) {
					if err := conn.SetKeepAlive(true); tr.IsOK(err) {
						if iface := tcpiface.New(conn); iface != nil {
							s.RemoteAddr = conn.RemoteAddr().String()
							s.Requester = requester.New(iface, s.Enigma)
							s.Responder = responder.New(iface, s.Enigma)
							retChan <- vtc.Ok
							return
						}
					}
				}
			}
		}
	}
	retChan <- vtc.Error
}

func (s *Stream) runClient(ctx context.Context, wg *sync.WaitGroup, retChan chan<- vtc.OperationStatusType) {
	defer wg.Done()

	rc := make(chan vtc.OperationStatusType)
	ictx, cancel := context.WithCancel(context.Background())
	var iwg sync.WaitGroup
	iwg.Add(1)

	go s.connectToServer(ictx, &iwg, rc)

	select {
	case <-ctx.Done():
		cancel()
		iwg.Wait()
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Stream.runClient: timeout")
			retChan <- vtc.Timeout
		} else {
			log.Println("Stream.runClient:", ctx.Err())
		}
	case retval := <-rc:
		iwg.Wait()
		retChan <- retval
	}
}

func (s *Stream) connectToServer(ctx context.Context, wg *sync.WaitGroup, retChan chan<- vtc.OperationStatusType) {
	defer wg.Done()

	if addr, err := net.ResolveIPAddr("ip", s.ServerAddr); tr.IsOK(err) {
		tcpAddr := net.TCPAddr{IP: addr.IP, Port: s.ServerPort, Zone: addr.Zone}
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if s.dial(tcpAddr) {
					log.Println("connection with server established")
					retChan <- vtc.Ok
					return
				}
				// następna próba za 5 sekund
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// Client
// Próba połączenia z serwerem.
func (s *Stream) dial(tcpAddr net.TCPAddr) bool {
	if conn, err := net.DialTCP("tcp", nil, &tcpAddr); tr.IsOK(err) {
		if err := conn.SetKeepAlive(true); tr.IsOK(err) {
			if iface := tcpiface.New(conn); tr.IsOK(err) {
				s.RemoteAddr = conn.RemoteAddr().String()
				s.Requester = requester.New(iface, s.Enigma)
				s.Responder = responder.New(iface, s.Enigma)
				return true
			}
		}
	}
	return false
}
