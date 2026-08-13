package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	authservice "github.com/accelolabs/avito-tamagochi/backend/internal/auth/service"
	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	leadservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/service"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	rewardsservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/service"
	summaryservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/service"
	tasksservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/service"
	"github.com/google/uuid"
)

type Handler struct {
	petService         petservice.Service
	tasksService       tasksservice.Service
	rewardsService     rewardsservice.Service
	summaryService     summaryservice.Service
	leaderboardService leadservice.Service
	authService        authservice.Service
}

func currentUserID(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserID(r.Context())
}

func New(petService petservice.Service, tasksService tasksservice.Service, rewardsService rewardsservice.Service, summaryService summaryservice.Service, leaderboardService leadservice.Service, authService authservice.Service) *Handler {
	return &Handler{petService: petService, tasksService: tasksService, rewardsService: rewardsService, summaryService: summaryService, leaderboardService: leaderboardService, authService: authService}
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
	handle(http.MethodGet, "/api/v1/leaderboard/week", h.getWeeklyLeaderboard)
	handle(http.MethodGet, "/api/v1/leaderboard/month", h.getMonthlyLeaderboard)
	handle(http.MethodGet, "/api/v1/leaderboard/streak", h.getStreakLeaderboard)
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
	case errors.Is(err, gameerrors.ErrTaskNotAvailable):
		writeError(w, http.StatusBadRequest, "task_not_available", "task is not available today")
	case errors.Is(err, gameerrors.ErrRewardNotFound):
		writeError(w, http.StatusNotFound, "reward_not_found", "reward was not found")
	case errors.Is(err, gameerrors.ErrRewardAlreadyUsed):
		writeError(w, http.StatusConflict, "reward_already_used", "reward was already used")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
