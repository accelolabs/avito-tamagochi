package handlers

import (
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
)

type registerRequest struct {
	Email       string
	password    string
	DisplayName string
}

func (request *registerRequest) UnmarshalJSON(data []byte) error {
	values, err := decodeStringObject(data, "email", "password", "displayName")
	if err != nil {
		return err
	}
	request.Email = values["email"]
	request.password = values["password"]
	request.DisplayName = values["displayName"]
	return nil
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	user, session, err := h.service.Register(r.Context(), model.RegisterInput{Email: request.Email, Password: request.password, DisplayName: request.DisplayName})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}
