package handlers

import (
	"net/http"

	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	"github.com/google/uuid"
)

type taskResponse struct {
	ID            uuid.UUID      `json:"id"`
	Type          taskmodel.Type `json:"type"`
	Progress      int            `json:"progress"`
	RequiredCount int            `json:"requiredCount"`
	Status        string         `json:"status"`
	XPReward      int            `json:"xpReward"`
}

type tasksResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

func (h *Handler) getTodayTasks(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

func (h *Handler) applyMockAction(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}
