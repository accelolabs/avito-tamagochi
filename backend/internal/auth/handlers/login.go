package handlers

import (
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
)

type loginRequest struct {
	Email    string
	password string
}

func (request *loginRequest) UnmarshalJSON(data []byte) error {
	values, err := decodeStringObject(data, "email", "password")
	if err != nil {
		return err
	}
	request.Email = values["email"]
	request.password = values["password"]
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	user, session, err := h.service.Login(r.Context(), model.LoginInput{Email: request.Email, Password: request.password})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusOK, toUserResponse(user))
}
