package handlers

import (
	"net/http"

	"github.com/google/uuid"
)

type leaderboardEntryResponse struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"userId"`
	DisplayName string    `json:"displayName"`
	XP          int       `json:"xp"`
	Level       int       `json:"level"`
}

type leaderboardResponse struct {
	Entries     []leaderboardEntryResponse `json:"entries"`
	CurrentUser *leaderboardEntryResponse  `json:"currentUser,omitempty"`
}

func (h *Handler) getLeaderboard(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}
