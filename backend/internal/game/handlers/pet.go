package handlers

import (
	"net/http"
	"time"

	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
)

type petResponse struct {
	XP            int            `json:"xp"`
	Level         int            `json:"level"`
	Stage         petmodel.Stage `json:"stage"`
	Energy        int            `json:"energy"`
	LastChargedAt time.Time      `json:"lastChargedAt"`
	IsDead        bool           `json:"isDead"`
}

type chargeResultResponse struct {
	Pet            petResponse `json:"pet"`
	BaseChargeXP   int         `json:"baseChargeXp"`
	DailyRewardXP  int         `json:"dailyRewardXp"`
	TotalXPAwarded int         `json:"totalXpAwarded"`
}

type petActionResultResponse struct {
	Pet       petResponse `json:"pet"`
	XPAwarded int         `json:"xpAwarded"`
}

type streakResponse struct {
	CurrentStreak     int     `json:"currentStreak"`
	LongestStreak     int     `json:"longestStreak"`
	LastChargeDate    *string `json:"lastChargeDate"`
	NextDailyRewardXP int     `json:"nextDailyRewardXp"`
}

func (h *Handler) getPet(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	stats, err := h.petService.GetPet(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPetResponse(stats))
}

func (h *Handler) getStreak(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	stats, err := h.petService.GetStreak(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	var lastChargeDate *string
	if stats.LastChargeDate != nil {
		value := stats.LastChargeDate.Format(time.DateOnly)
		lastChargeDate = &value
	}
	writeJSON(w, http.StatusOK, streakResponse{
		CurrentStreak: stats.CurrentStreak, LongestStreak: stats.LongestStreak,
		LastChargeDate: lastChargeDate, NextDailyRewardXP: stats.NextDailyRewardXP,
	})
}

func (h *Handler) applyPetAction(w http.ResponseWriter, r *http.Request) {
	var action string
	if !decodeJSONBody(w, r, &action, 64) {
		return
	}
	if action != "charge" && action != "pet" {
		writeError(w, http.StatusBadRequest, "validation_error", "request is invalid")
		return
	}
	id, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	switch action {
	case "charge":
		result, err := h.petService.ChargePet(r.Context(), id)
		if err != nil {
			mapGameError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, chargeResultResponse{
			Pet: toPetResponse(result.Pet), BaseChargeXP: result.BaseChargeXP,
			DailyRewardXP: result.DailyRewardXP, TotalXPAwarded: result.TotalXPAwarded,
		})
	case "pet":
		result, err := h.petService.Pet(r.Context(), id)
		if err != nil {
			mapGameError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, petActionResultResponse{Pet: toPetResponse(result.Pet), XPAwarded: result.XPAwarded})
	default:
		writeError(w, http.StatusBadRequest, "validation_error", "request is invalid")
	}
}

func toPetResponse(stats *petmodel.Stats) petResponse {
	return petResponse{
		XP: stats.XP, Level: stats.Level, Stage: stats.Stage, Energy: stats.Energy,
		LastChargedAt: stats.LastChargedAt, IsDead: stats.IsDead,
	}
}
