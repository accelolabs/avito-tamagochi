package handlers

import (
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		if sessionID, parseErr := uuid.Parse(cookie.Value); parseErr == nil {
			if logoutErr := h.service.Logout(r.Context(), sessionID); logoutErr != nil {
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				return
			}
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
