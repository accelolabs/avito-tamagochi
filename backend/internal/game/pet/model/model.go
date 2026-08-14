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
	ID                uuid.UUID
	UserID            uuid.UUID
	XP                int
	EnergyPercent     int
	EnergyUpdatedAt   time.Time
	LastChargedAt     time.Time
	ChargeStreak      int
	LongestStreak     int
	LastStreakDate    *time.Time
	StreakStartedDate *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewPet(userID uuid.UUID, now time.Time) Pet {
	return Pet{
		ID: uuid.New(), UserID: userID, EnergyPercent: 50, EnergyUpdatedAt: now,
		LastChargedAt: now.Add(-24 * time.Hour), CreatedAt: now, UpdatedAt: now,
	}
}

type Stats struct {
	XP            int
	Level         int
	Stage         Stage
	Energy        int
	LastChargedAt time.Time
	IsDead        bool
}

type ChargeResult struct {
	Pet            *Stats
	BaseChargeXP   int
	DailyRewardXP  int
	TotalXPAwarded int
}

type PetActionResult struct {
	Pet       *Stats
	XPAwarded int
}

type StreakStats struct {
	CurrentStreak     int
	LongestStreak     int
	LastChargeDate    *time.Time
	NextDailyRewardXP int
}
