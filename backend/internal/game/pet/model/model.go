package model

import (
	"time"

	"github.com/google/uuid"
)

type Stage string

const (
	Egg   Stage = "egg"
	Child Stage = "child"
	Teen  Stage = "teen"
	Adult Stage = "adult"
)

type Pet struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	XP             int
	LastChargedAt  time.Time
	ChargeStreak   int
	LastStreakDate  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Stats struct {
	XP            int
	Level         int
	Stage         Stage
	Energy        int
	LastChargedAt time.Time
	ChargeStreak  int
	IsDead        bool
}

