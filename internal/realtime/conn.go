package realtime

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

// Conn adapts a coder/websocket connection into a Hub Subscriber. All writes
// go through a single buffered channel drained by one writer goroutine, so
// concurrent Broadcast calls from other connections' read loops never race
// on the underlying socket and never block the broadcaster.
//
// A slow reader is disconnected rather than allowed to backpressure the
// whole topic — see Send.
type Conn struct {
	ws   *websocket.Conn
	send chan Message
	done chan struct{}
	once sync.Once
}

const outboundBuffer = 64

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{
		ws:   ws,
		send: make(chan Message, outboundBuffer),
		done: make(chan struct{}),
	}
}

// Send enqueues msg for delivery. If the connection's outbound buffer is
// full, the connection is closed instead of dropping or blocking: losing a
// document update silently would corrupt collaborative state, so a
// connection that cannot keep up must resync over a fresh connection.
func (c *Conn) Send(msg Message) {
	select {
	case c.send <- msg:
	case <-c.done:
	default:
		c.Close()
	}
}

// Close closes the connection exactly once. Safe to call multiple times and
// from multiple goroutines (Hub.Shutdown, the read pump on error, and the
// handler's deferred cleanup all call it).
func (c *Conn) Close() {
	c.once.Do(func() {
		close(c.done)
		_ = c.ws.CloseNow()
	})
}

// WritePump drains the outbound queue until the connection closes or ctx is
// canceled. Run it in its own goroutine; it returns when done.
func (c *Conn) WritePump(ctx context.Context) {
	for {
		select {
		case msg := <-c.send:
			typ := websocket.MessageText
			if msg.Kind == Binary {
				typ = websocket.MessageBinary
			}
			if err := c.ws.Write(ctx, typ, msg.Payload); err != nil {
				c.Close()
				return
			}
		case <-c.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Done is closed once the connection has been closed.
func (c *Conn) Done() <-chan struct{} {
	return c.done
}
