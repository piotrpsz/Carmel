package session

import (
	"Carmel/connector/stream"
	"context"
	"github.com/gotk3/gotk3/gtk"
)

type Session struct {
	In  *stream.Stream
	Out *stream.Stream
}

func ServerNew(app *gtk.Application, ctx context.Context, port int) *Session {
	return &Session{In: stream.Server(port), Out: stream.Server(port + 1)}
}

func ClientNew(app *gtk.Application, ctx context.Context, addr, name string, port, timeout int) *Session {
	return &Session{In: stream.Client(addr, port, 120), Out: stream.Client(addr, port+1, 120)}
}
