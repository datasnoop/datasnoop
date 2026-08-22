// Package live provides best-effort post-commit SSE delivery.
package live

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}
type Broker struct {
	mu       sync.Mutex
	clients  map[chan Event]struct{}
	capacity int
}

func NewBroker(capacity int) *Broker {
	return &Broker{clients: map[chan Event]struct{}{}, capacity: capacity}
}
func (b *Broker) ClientCount() int { b.mu.Lock(); defer b.mu.Unlock(); return len(b.clients) }
func (b *Broker) Publish(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for client := range b.clients {
		select {
		case client <- event:
		default:
			select {
			case <-client:
			default:
			}
			select {
			case client <- Event{Type: "gap", Data: "live events were dropped; refresh history"}:
			default:
			}
		}
	}
}
func (b *Broker) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", 500)
			return
		}
		client := make(chan Event, b.capacity)
		b.mu.Lock()
		b.clients[client] = struct{}{}
		b.mu.Unlock()
		defer func() { b.mu.Lock(); delete(b.clients, client); b.mu.Unlock() }()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()
		for {
			select {
			case event := <-client:
				payload, _ := json.Marshal(event)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})
}
