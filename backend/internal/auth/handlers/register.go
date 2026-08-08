package handlers

import (
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	user, session, err := h.service.Register(r.Context(), model.RegisterInput{Email: request.Email, Password: request.Password, DisplayName: request.DisplayName})
	if err != nil {
		mapServiceError(w, err, false)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}
