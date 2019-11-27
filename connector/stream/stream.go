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
	enigma     *enigma.Enigma
	Responder  *responder.Responder
	Requester  *requester.Requester
	serverAddr string // client only
	serverPort int    // server & client
	timeout    int    // client only
}

func Server(port int) *Stream {
	if e := enigma.New(); e != nil {
		s := &Stream{role: vtc.Server, serverPort: port}
		if s.initKeys() {
			return s
		}
	}
	return nil
}

func (s *Stream) initKeys() bool {
	if key := secret.RandomBytes(blowfish.MaxKeyLength); s.enigma.InitBlowfish(key) {
		if key := secret.RandomBytes(ghost.KeySize); s.enigma.InitGhost(key) {
			if key := secret.RandomBytes(way3.KeySize); s.enigma.InitWay3(key) {
				return true
			}
		}
	}
	return false
}

func Client(addr string, port, timeout int) *Stream {
	if e := enigma.New(); e != nil {
		return &Stream{role: vtc.Client, serverAddr: addr, serverPort: port, enigma: e, timeout: timeout}
	}
	return nil
}

/********************************************************************
*                                                                   *
*                             R U N                                 *
*                                                                   *
********************************************************************/

func (s *Stream) Run(ctx context.Context, wg *sync.WaitGroup) bool {
	defer wg.Done()

	retChan := make(chan bool)
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
			return false
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
			return false
		case retv := <-retChan:
			iwg.Wait()
			return retv
		}

	}
	return false
}

func (s *Stream) runServer(ctx context.Context, wg *sync.WaitGroup, retChan chan<- bool) {
	defer wg.Done()

	rc := make(chan bool)
	go s.waitForClient(rc)

	select {
	case <-ctx.Done():
		log.Println("Stream.runServer:", ctx.Err())
	case retv := <-rc:
		retChan <- retv
	}
}

func (s *Stream) waitForClient(retChan chan<- bool) {
	if hostname, err := os.Hostname(); tr.IsOK(err) {
		if addr, err := net.ResolveIPAddr("ip", hostname); tr.IsOK(err) {
			tcpAdrr := net.TCPAddr{IP: addr.IP, Port: s.serverPort, Zone: addr.Zone}
			fmt.Printf("Server: %s:%d\n", tcpAdrr.IP, tcpAdrr.Port)

			if listener, err := net.ListenTCP("tcp", &tcpAdrr); tr.IsOK(err) {
				if conn, err := listener.AcceptTCP(); tr.IsOK(err) {
					if err := conn.SetKeepAlive(true); tr.IsOK(err) {
						if iface := tcpiface.New(conn); iface != nil {
							if s.enigma.InitConnection(iface, vtc.Server) {
								s.Requester = requester.New(iface, s.enigma)
								s.Responder = responder.New(iface, s.enigma)
								retChan <- true
								return
							}
						}
					}
				}
			}
		}
	}
	retChan <- false
}

func (s *Stream) runClient(ctx context.Context, wg *sync.WaitGroup, retChan chan<- bool) {
	defer wg.Done()

	rc := make(chan bool)
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
			retChan <- false
		} else {
			log.Println("Stream.runClient:", ctx.Err())
		}
	case retval := <-rc:
		iwg.Wait()
		retChan <- retval
	}
}

func (s *Stream) connectToServer(ctx context.Context, wg *sync.WaitGroup, retChan chan<- bool) {
	defer wg.Done()

	if addr, err := net.ResolveIPAddr("ip", s.serverAddr); tr.IsOK(err) {
		tcpAddr := net.TCPAddr{IP: addr.IP, Port: s.serverPort, Zone: addr.Zone}
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if s.dial(tcpAddr) {
					log.Println("connection with server established")
					retChan <- true
					return
				}
				// następna próba za 5 sekund
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (s *Stream) dial(tcpAddr net.TCPAddr) bool {
	if conn, err := net.DialTCP("tcp", nil, &tcpAddr); tr.IsOK(err) {
		if err := conn.SetKeepAlive(true); tr.IsOK(err) {
			if iface := tcpiface.New(conn); tr.IsOK(err) {
				if s.enigma.InitConnection(iface, vtc.Client) {
					s.Requester = requester.New(iface, s.enigma)
					s.Responder = responder.New(iface, s.enigma)
					return true
				}
			}
		}
	}
	return false
}
