package pet

import (
	"time"

	"github.com/google/uuid"
)

type Pet struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	XP            int
	LastChargedAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Stats struct {
	XP            int       `json:"xp"`
	Level         int       `json:"level"`
	Energy        int       `json:"energy"`
	LastChargedAt time.Time `json:"lastChargedAt"`
}
