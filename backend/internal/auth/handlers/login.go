package handlers

import (
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	user, session, err := h.service.Login(r.Context(), auth.LoginInput{Email: request.Email, Password: request.Password})
	if err != nil {
		mapServiceError(w, err, true)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusOK, toUserResponse(user))
}
