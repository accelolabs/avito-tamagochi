package realtime

import (
	"encoding/json"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWS upgrades an authenticated HTTP request to a WebSocket connection.
func ServeWS(hub *Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:    hub,
			userID: userID,
			conn:   conn,
			send:   make(chan string, 256),
		}

		client.hub.register <- client
		go client.writePump()
		go client.readPump()
	})
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Code: code, Message: message})
}
