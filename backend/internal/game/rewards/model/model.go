package model

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	Promotion Type = "promotion"
	Delivery  Type = "free_delivery"
)

type UserReward struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	RewardType Type
	Status     string
	UnlockedAt time.Time
	UsedAt     *time.Time
}
