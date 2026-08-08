package handlers

import (
	"net/http"
	"time"

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

func (h *Handler) getRewards(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

func (h *Handler) useReward(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}
