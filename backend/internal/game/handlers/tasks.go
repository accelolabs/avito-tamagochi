package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	"github.com/google/uuid"
)

type taskResponse struct {
	ID            uuid.UUID      `json:"id,omitempty"`
	Type          taskmodel.Type `json:"type"`
	Progress      int            `json:"progress"`
	RequiredCount int            `json:"requiredCount"`
	Status        string         `json:"status"`
	XPReward      int            `json:"xpReward"`
}

type tasksResponse struct {
	Tasks []taskResponse `json:"tasks"`
}

func (h *Handler) getTodayTasks(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, 401, "unauthorized", "authentication is required")
		return
	}
	items, err := h.tasksService.GetTodayTasks(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	result := tasksResponse{Tasks: make([]taskResponse, 0, len(items))}
	for _, item := range items {
		status := "in_progress"
		if item.Completed {
			status = "completed"
		}
		result.Tasks = append(result.Tasks, taskResponse{ID: item.ID, Type: item.TaskType, Progress: item.Progress, RequiredCount: item.RequiredCount, Status: status, XPReward: item.XPReward})
	}
	writeJSON(w, 200, result)
}

func (h *Handler) applyMockAction(w http.ResponseWriter, r *http.Request) {
	var action taskmodel.Type
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64)).Decode(&action); err != nil {
		writeError(w, 400, "validation_error", "request is invalid")
		return
	}
	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, 401, "unauthorized", "authentication is required")
		return
	}
	if err := h.tasksService.ApplyAction(r.Context(), id, action); err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "applied"})
}
