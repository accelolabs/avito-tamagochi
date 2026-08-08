package model

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
	XP            int
	Level         int
	Energy        int
	LastChargedAt time.Time
}
