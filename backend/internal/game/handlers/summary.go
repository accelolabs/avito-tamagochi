package handlers

import "net/http"

type summaryResponse struct {
	XPEarned       int      `json:"xpEarned"`
	CompletedTasks int      `json:"completedTasks"`
	Charges        int      `json:"charges"`
	Level          int      `json:"level"`
	UnlockedReward []string `json:"unlockedRewards"`
}

func (h *Handler) getTodaySummary(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}
