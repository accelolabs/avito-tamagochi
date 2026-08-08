package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/google/uuid"
)

type contextKey struct{}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return id, ok
}

func RequireSession(service auth.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			unauthorized(w)
			return
		}
		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			unauthorized(w)
			return
		}
		userID, err := service.ValidateSession(r.Context(), sessionID)
		if errors.Is(err, auth.ErrSessionNotFound) {
			unauthorized(w)
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, userID)))
	})
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Code: code, Message: message})
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
