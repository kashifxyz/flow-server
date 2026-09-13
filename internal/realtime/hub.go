// Package realtime provides a minimal in-process publish/subscribe hub for
// WebSocket fan-out (presence and document-update relay). It intentionally
// knows nothing about HTTP, Yjs, or presence semantics — those live in the
// modules that use it (see internal/modules/collaboration).
//
// This is single-instance only. If Flow ever runs multiple API instances,
// swap the Hub's storage for a Redis-backed pub/sub implementation behind
// the same Subscribe/Unsubscribe/Broadcast shape (PROJECT-INFO.md §37) —
// callers do not need to change.
package realtime

import "sync"

type MessageKind int

const (
	Text MessageKind = iota
	Binary
)

type Message struct {
	Kind    MessageKind
	Payload []byte
}

// Subscriber receives messages broadcast to a topic. Send must never block
// the caller (a slow or dead subscriber should disconnect itself, not stall
// the broadcaster) — see Conn.Send in ws.go for the reference implementation.
type Subscriber interface {
	Send(Message)
}

type Hub struct {
	mu     sync.RWMutex
	topics map[string]map[Subscriber]struct{}
	all    map[Subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{
		topics: make(map[string]map[Subscriber]struct{}),
		all:    make(map[Subscriber]struct{}),
	}
}

func (h *Hub) Subscribe(topic string, sub Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs, ok := h.topics[topic]
	if !ok {
		subs = make(map[Subscriber]struct{})
		h.topics[topic] = subs
	}
	subs[sub] = struct{}{}
	h.all[sub] = struct{}{}
}

// Unsubscribe removes sub from topic. Callers that hold a connection across
// multiple topics must call this once per topic and rely on Forget to drop
// it from the global registry once fully done.
func (h *Hub) Unsubscribe(topic string, sub Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs, ok := h.topics[topic]
	if !ok {
		return
	}
	delete(subs, sub)
	if len(subs) == 0 {
		delete(h.topics, topic)
	}
}

// Forget drops sub from the hub's global registry (used by Shutdown). Call
// it once a connection has closed and unsubscribed from all of its topics.
func (h *Hub) Forget(sub Subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.all, sub)
}

func (h *Hub) Broadcast(topic string, msg Message, except Subscriber) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.topics[topic] {
		if sub == except {
			continue
		}
		sub.Send(msg)
	}
}

func (h *Hub) Count(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.topics[topic])
}

// Closer is implemented by subscribers that can be told to close their
// underlying transport (every *Conn from ws.go). Shutdown uses it to give
// live WebSocket connections a clean close frame instead of letting the
// process exit drop them.
type Closer interface {
	Close()
}

// Shutdown closes every currently registered subscriber that implements
// Closer. Call it from the server's graceful-shutdown path.
func (h *Hub) Shutdown() {
	h.mu.RLock()
	subs := make([]Subscriber, 0, len(h.all))
	for sub := range h.all {
		subs = append(subs, sub)
	}
	h.mu.RUnlock()
	for _, sub := range subs {
		if c, ok := sub.(Closer); ok {
			c.Close()
		}
	}
}
