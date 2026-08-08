package handlers

import (
	"net/http"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	"github.com/google/uuid"
)

type rewardResponse struct {
	ID         uuid.UUID        `json:"id"`
	Type       rewardmodel.Type `json:"type"`
	Status     string           `json:"status"`
	UnlockedAt time.Time        `json:"unlockedAt"`
	UsedAt     *time.Time       `json:"usedAt,omitempty"`
}

type rewardsResponse struct {
	Rewards []rewardResponse `json:"rewards"`
}

func (h *Handler) getRewards(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, 401, "unauthorized", "authentication is required")
		return
	}
	items, err := h.rewardsService.GetRewards(r.Context(), id)
	if err != nil {
		mapGameError(w, err)
		return
	}
	result := rewardsResponse{Rewards: make([]rewardResponse, 0, len(items))}
	for _, item := range items {
		result.Rewards = append(result.Rewards, rewardResponse{ID: item.ID, Type: item.RewardType, Status: item.Status, UnlockedAt: item.UnlockedAt, UsedAt: item.UsedAt})
	}
	writeJSON(w, 200, result)
}

func (h *Handler) useReward(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, 401, "unauthorized", "authentication is required")
		return
	}
	rewardID, err := uuid.Parse(r.PathValue("rewardID"))
	if err != nil {
		writeError(w, 400, "validation_error", "request is invalid")
		return
	}
	item, err := h.rewardsService.UseReward(r.Context(), id, rewardID)
	if err != nil {
		mapGameError(w, err)
		return
	}
	writeJSON(w, 200, rewardResponse{ID: item.ID, Type: item.RewardType, Status: item.Status, UnlockedAt: item.UnlockedAt, UsedAt: item.UsedAt})
}
