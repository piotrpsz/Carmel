package session

import (
	"Carmel/connector/stream"
)

type Session struct {
	In  *stream.Stream
	Out *stream.Stream
}

func ServerNew(port int) *Session {
	return &Session{In: stream.Server(port), Out: stream.Server(port + 1)}
}

func ClientNew(addr string, port, timeout int) *Session {
	return &Session{In: stream.Client(addr, port, timeout), Out: stream.Client(addr, port+1, timeout)}
}
