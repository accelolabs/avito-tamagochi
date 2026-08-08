package realtime

import (
	"github.com/google/uuid"
)

// EventNotifier - интерфейс, который мы будем передавать другим сервисам (Pet, Tasks),
// чтобы они могли отправлять уведомления, ничего не зная про вебсокеты.
type EventNotifier interface {
	NotifyUser(userID uuid.UUID, eventType string)
}

// Notification описывает структуру тонкого события
type Notification struct {
	UserID    uuid.UUID
	EventType string
}

type Hub struct {
	// clients: userID -> набор активных подключений (вкладок)
	clients map[uuid.UUID]map[*Client]bool

	// Канал для отправки событий конкретным юзерам
	notify chan Notification

	// Регистрация новых подключений
	register chan *Client

	// Удаление отключившихся клиентов
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

// Run запускается в отдельной горутине при старте сервера
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
					// Если закрыли последнюю вкладку — удаляем юзера из мапы
					if len(connections) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}

		case notification := <-h.notify:
			// Ищем все активные подключения этого юзера и рассылаем пинг
			if connections, ok := h.clients[notification.UserID]; ok {
				for client := range connections {
					select {
					case client.send <- notification.EventType:
						// Успешно отправлено
					default:
						// Канал забит, клиент завис — отключаем
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

// NotifyUser реализует интерфейс EventNotifier для других модулей
func (h *Hub) NotifyUser(userID uuid.UUID, eventType string) {
	h.notify <- Notification{
		UserID:    userID,
		EventType: eventType,
	}
}
