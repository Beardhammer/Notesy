package shout

import (
	"testing"
	"time"
)

func TestHubBroadcastDeliversToSubscribers(t *testing.T) {
	h := NewHub()
	a := h.Subscribe()
	b := h.Subscribe()
	defer h.Unsubscribe(a)
	defer h.Unsubscribe(b)

	msg := &Message{ID: 1, Body: "hi"}
	h.Broadcast(msg)

	for _, c := range []*Client{a, b} {
		select {
		case got := <-c.Ch:
			if got.ID != 1 {
				t.Fatalf("wrong ID: %d", got.ID)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("client did not receive broadcast")
		}
	}
}

func TestHubBroadcastDoesNotBlockOnFullBuffer(t *testing.T) {
	h := NewHub()
	c := h.Subscribe()
	defer h.Unsubscribe(c)

	// Fill the buffer without draining, then broadcast many more.
	// Must not hang.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			h.Broadcast(&Message{ID: uint(i)})
		}
		close(done)
	}()

	select {
	case <-done:
		// pass
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("Broadcast blocked on slow client")
	}
}

func TestHubUnsubscribeDuringBroadcastDoesNotPanic(t *testing.T) {
	h := NewHub()
	c := h.Subscribe()
	// Concurrently unsubscribe while broadcasting.
	go func() {
		for i := 0; i < 100; i++ {
			h.Broadcast(&Message{ID: uint(i)})
		}
	}()
	for i := 0; i < 10; i++ {
		h.Unsubscribe(c)
		c = h.Subscribe()
	}
	h.Unsubscribe(c)
	// If we reach here without panic, pass.
}
