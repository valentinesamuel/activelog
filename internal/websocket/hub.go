package websocket

import (
	"sync"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients    map[int]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub event loop. It must be run in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Client buffer full; drop message
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser delivers a message to a specific connected user
func (h *Hub) SendToUser(userID int, msgType string, payload any) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	msg := Message{
		Type:      msgType,
		UserID:    userID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	select {
	case client.send <- msg:
	default:
		// Client buffer full; drop message
	}
}
