// shout/hub.go
package shout

import "sync"

// Client is a single SSE subscriber.
type Client struct {
	Ch chan *Message // buffered
}

// Hub fans out messages to subscribed SSE clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

// Subscribe registers a new client with a 16-deep buffer.
func (h *Hub) Subscribe() *Client {
	c := &Client{Ch: make(chan *Message, 16)}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	return c
}

// Unsubscribe removes a client and closes its channel.
// Safe to call multiple times.
func (h *Hub) Unsubscribe(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Ch)
	}
	h.mu.Unlock()
}

// Broadcast sends m to every client. Slow clients (buffer full) drop the message.
func (h *Hub) Broadcast(m *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.Ch <- m:
		default:
			// buffer full — drop; client will backfill via Last-Event-ID on reconnect
		}
	}
}
