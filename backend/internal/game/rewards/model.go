package rewards

import (
	"time"

	"github.com/google/uuid"
)

type RewardType string

const (
	RewardPromotion RewardType = "promotion"
	RewardDelivery  RewardType = "free_delivery"
)

type UserReward struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	RewardType RewardType
	Status     string
	UnlockedAt time.Time
	UsedAt     *time.Time
}
