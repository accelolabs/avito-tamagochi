package handlers

import (
	"context"
	"net/http"
	"time"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type leaderboardEntryResponse struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"userId"`
	DisplayName string    `json:"displayName"`
	XP          int       `json:"xp"`
	Level       int       `json:"level"`
}

type deltaLeaderboardEntryResponse struct {
	leaderboardEntryResponse
	XPDelta int `json:"xpDelta"`
}

type leaderboardResponse[T any] struct {
	Entries     []T `json:"entries"`
	CurrentUser *T  `json:"currentUser"`
}

type streakLeaderboardEntryResponse struct {
	Rank            int       `json:"rank"`
	UserID          uuid.UUID `json:"userId"`
	DisplayName     string    `json:"displayName"`
	CurrentStreak   int       `json:"currentStreak"`
	LongestStreak   int       `json:"longestStreak"`
	StreakStartedAt string    `json:"streakStartedAt"`
	LastChargeDate  string    `json:"lastChargeDate"`
}

func (h *Handler) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	board, err := h.leaderboardService.GetAll(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapXPBoard(board, false))
}

func (h *Handler) getWeeklyLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.getDeltaLeaderboard(w, r, h.leaderboardService.GetWeekly)
}

func (h *Handler) getMonthlyLeaderboard(w http.ResponseWriter, r *http.Request) {
	h.getDeltaLeaderboard(w, r, h.leaderboardService.GetMonthly)
}

func (h *Handler) getDeltaLeaderboard(
	w http.ResponseWriter,
	r *http.Request,
	load func(context.Context, uuid.UUID) (*leadermodel.XPBoard, error),
) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	board, err := load(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapXPBoard(board, true))
}

func (h *Handler) getStreakLeaderboard(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	board, err := h.leaderboardService.GetStreak(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	entries := make([]streakLeaderboardEntryResponse, 0, len(board.Entries))
	for i := range board.Entries {
		entries = append(entries, toStreakLeaderboardResponse(&board.Entries[i]))
	}
	var current *streakLeaderboardEntryResponse
	if board.CurrentUser != nil {
		value := toStreakLeaderboardResponse(board.CurrentUser)
		current = &value
	}
	writeJSON(w, http.StatusOK, leaderboardResponse[streakLeaderboardEntryResponse]{
		Entries: entries, CurrentUser: current,
	})
}

func mapXPBoard(board *leadermodel.XPBoard, includeDelta bool) any {
	if includeDelta {
		entries := make([]deltaLeaderboardEntryResponse, 0, len(board.Entries))
		for i := range board.Entries {
			entries = append(entries, toDeltaLeaderboardResponse(&board.Entries[i]))
		}
		var current *deltaLeaderboardEntryResponse
		if board.CurrentUser != nil {
			value := toDeltaLeaderboardResponse(board.CurrentUser)
			current = &value
		}
		return leaderboardResponse[deltaLeaderboardEntryResponse]{Entries: entries, CurrentUser: current}
	}

	entries := make([]leaderboardEntryResponse, 0, len(board.Entries))
	for i := range board.Entries {
		entries = append(entries, toLeaderboardResponse(&board.Entries[i]))
	}
	var current *leaderboardEntryResponse
	if board.CurrentUser != nil {
		value := toLeaderboardResponse(board.CurrentUser)
		current = &value
	}
	return leaderboardResponse[leaderboardEntryResponse]{Entries: entries, CurrentUser: current}
}

func toLeaderboardResponse(value *leadermodel.XPEntry) leaderboardEntryResponse {
	return leaderboardEntryResponse{
		Rank: value.Rank, UserID: value.UserID, DisplayName: value.DisplayName,
		XP: value.XP, Level: value.Level,
	}
}

func toDeltaLeaderboardResponse(value *leadermodel.XPEntry) deltaLeaderboardEntryResponse {
	return deltaLeaderboardEntryResponse{
		leaderboardEntryResponse: toLeaderboardResponse(value), XPDelta: value.XPDelta,
	}
}

func toStreakLeaderboardResponse(value *leadermodel.StreakEntry) streakLeaderboardEntryResponse {
	return streakLeaderboardEntryResponse{
		Rank:            value.Rank,
		UserID:          value.UserID,
		DisplayName:     value.DisplayName,
		CurrentStreak:   value.CurrentStreak,
		LongestStreak:   value.LongestStreak,
		StreakStartedAt: value.StreakStartedDate.Format(time.DateOnly),
		LastChargeDate:  value.LastChargeDate.Format(time.DateOnly),
	}
}
