package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	game "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
)

type Handler struct {
	petService  game.Service
	authService auth.Service
}

func New(petService game.Service, authService auth.Service) *Handler {
	return &Handler{petService: petService, authService: authService}
}

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	handle := func(method, pattern string, handler http.HandlerFunc) {
		mux.Handle(method+" "+pattern, middleware.RequireSession(h.authService, handler))
	}

	handle(http.MethodGet, "/api/v1/pet", h.getPet)
	handle(http.MethodPost, "/api/v1/pet/actions", h.chargePet)
	handle(http.MethodGet, "/api/v1/tasks/today", h.getTodayTasks)
	handle(http.MethodPost, "/api/v1/mock-avito/actions", h.applyMockAction)
	handle(http.MethodGet, "/api/v1/rewards", h.getRewards)
	handle(http.MethodPost, "/api/v1/rewards/{rewardID}/use", h.useReward)
	handle(http.MethodGet, "/api/v1/summary/today", h.getTodaySummary)
	handle(http.MethodGet, "/api/v1/leaderboard", h.getLeaderboard)
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func mapGameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gameerrors.ErrInvalidAction):
		writeError(w, http.StatusBadRequest, "validation_error", "request is invalid")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "not_implemented", "endpoint is not implemented yet")
}
