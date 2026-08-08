package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Для MVP и хакатона разрешаем любые CORS запросы,
	// так как Nginx сам всё смаршрутизирует
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS — хендлер для Gin, ожидается, что он стоит ПОСЛЕ Auth Middleware
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Достаем ID пользователя из контекста (положил Auth Middleware)
		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, ok := userIDRaw.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		// Обновляем HTTP до WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// Ошибку логирует сам Upgrader
			return
		}

		client := &Client{
			hub:    hub,
			userID: userID,
			conn:   conn,
			send:   make(chan string, 256), // Буфер на 256 непрочитанных событий
		}

		// Регистрируем клиента в Хабе
		client.hub.register <- client

		// Запускаем воркеры в отдельных горутинах
		go client.writePump()
		go client.readPump()
	}
}
