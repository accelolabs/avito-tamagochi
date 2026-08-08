package realtime

import (
	"github.com/google/uuid"
)

// EventNotifier lets business services publish user-scoped events without
// depending on the WebSocket implementation.
type EventNotifier interface {
	NotifyUser(userID uuid.UUID, eventType string)
}

type Notification struct {
	UserID    uuid.UUID
	EventType string
}

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	notify     chan Notification
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		notify:     make(chan Notification),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run processes connection lifecycle and notification events.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true

		case client := <-h.unregister:
			if connections, ok := h.clients[client.userID]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
					close(client.send)
					if len(connections) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}

		case notification := <-h.notify:
			if connections, ok := h.clients[notification.UserID]; ok {
				for client := range connections {
					select {
					case client.send <- notification.EventType:
					default:
						close(client.send)
						delete(connections, client)
					}
				}
				if len(connections) == 0 {
					delete(h.clients, notification.UserID)
				}
			}
		}
	}
}

func (h *Hub) NotifyUser(userID uuid.UUID, eventType string) {
	h.notify <- Notification{
		UserID:    userID,
		EventType: eventType,
	}
}
