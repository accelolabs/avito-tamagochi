package handlers

import (
	"net/http"
)

type summaryResponse struct {
	XPEarned       int      `json:"xpEarned"`
	CompletedTasks int      `json:"completedTasks"`
	Charges        int      `json:"charges"`
	Level          int      `json:"level"`
	CurrentXP      int      `json:"currentXP"`
	Energy         int      `json:"energy"`
	UnlockedRewards []string `json:"unlockedRewards"`
}

func (h *Handler) getTodaySummary(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	item, err := h.summaryService.GetToday(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	unlocked := item.UnlockedRewards
	if unlocked == nil {
		unlocked = []string{}
	}
	writeJSON(w, http.StatusOK, summaryResponse{
		XPEarned:       item.XPEarned,
		CompletedTasks: item.CompletedTasks,
		Charges:        item.Charges,
		Level:          item.Level,
		CurrentXP:      item.CurrentXP,
		Energy:         item.Energy,
		UnlockedRewards: unlocked,
	})
}
