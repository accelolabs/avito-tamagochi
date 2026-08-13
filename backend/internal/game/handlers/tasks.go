package handlers

import (
	"net/http"

	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
)

type taskResponse struct {
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
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
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
		result.Tasks = append(result.Tasks, taskResponse{Type: item.TaskType, Progress: item.Progress, RequiredCount: item.RequiredCount, Status: status, XPReward: item.XPReward})
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) applyMockAction(w http.ResponseWriter, r *http.Request) {
	var action taskmodel.Type
	if !decodeJSONBody(w, r, &action, 64) {
		return
	}
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	if err := h.tasksService.ApplyAction(r.Context(), id, action); err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}
