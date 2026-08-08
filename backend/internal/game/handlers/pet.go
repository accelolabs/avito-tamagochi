package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
)

type petResponse struct {
	XP            int       `json:"xp"`
	Level         int       `json:"level"`
	Energy        int       `json:"energy"`
	LastChargedAt time.Time `json:"lastChargedAt"`
}

func (h *Handler) getPet(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	stats, err := h.petService.GetPet(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPetResponse(stats))
}

func (h *Handler) chargePet(w http.ResponseWriter, r *http.Request) {
	var action string
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64))
	if err := decoder.Decode(&action); err != nil || action != "charge" {
		writeError(w, http.StatusBadRequest, "validation_error", "request is invalid")
		return
	}

	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	stats, err := h.petService.ChargePet(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPetResponse(stats))
}

func toPetResponse(stats *petmodel.Stats) petResponse {
	return petResponse{
		XP:            stats.XP,
		Level:         stats.Level,
		Energy:        stats.Energy,
		LastChargedAt: stats.LastChargedAt,
	}
}
