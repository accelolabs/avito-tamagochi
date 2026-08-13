package handlers

import (
	"net/http"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type leaderboardEntryResponse struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"userId"`
	DisplayName string    `json:"displayName"`
	XP          int       `json:"xp"`
	Level       int       `json:"level"`
	XPDelta     int       `json:"xpDelta"`
}
type leaderboardResponse struct {
	Entries     []leaderboardEntryResponse `json:"entries"`
	CurrentUser *leaderboardEntryResponse  `json:"currentUser"`
}

func (h *Handler) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all"
	}
	items, err := h.leaderboardService.GetTopByPeriod(r.Context(), period)
	if err != nil {
		mapGameError(w, err)
		return
	}
	current, err := h.leaderboardService.GetUserRankByPeriod(r.Context(), id, period)
	if err != nil {
		mapGameError(w, err)
		return
	}
	result := leaderboardResponse{Entries: make([]leaderboardEntryResponse, 0, len(items)), CurrentUser: toLeaderboardResponse(current)}
	for i := range items {
		result.Entries = append(result.Entries, *toLeaderboardResponse(&items[i]))
	}
	writeJSON(w, http.StatusOK, result)
}

func toLeaderboardResponse(value *leadermodel.Entry) *leaderboardEntryResponse {
	if value == nil {
		return nil
	}
	return &leaderboardEntryResponse{Rank: value.Rank, UserID: value.UserID, DisplayName: value.DisplayName, XP: value.XP, Level: value.Level, XPDelta: value.XPDelta}
}
