package ws

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains connected clients and broadcasts messages to all of them.
type Hub struct {
	clients    map[*client]struct{}
	mu         sync.RWMutex
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*client]struct{}),
		register:   make(chan *client, 16),
		unregister: make(chan *client, 16),
		broadcast:  make(chan []byte, 64),
	}
}

// Run processes register/unregister/broadcast events. Must be run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// Slow client — drop message rather than block
					log.Printf("ws: slow client, dropping message")
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast serialises event to JSON and fans it out to all connected clients.
func (h *Hub) Broadcast(event ActivityEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws: failed to marshal event: %v", err)
		return
	}
	h.broadcast <- data
}
