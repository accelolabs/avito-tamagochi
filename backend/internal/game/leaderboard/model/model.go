package model

import (
	"time"

	"github.com/google/uuid"
)

type XPEntry struct {
	Rank        int
	UserID      uuid.UUID
	DisplayName string
	XP          int
	Level       int
	XPDelta     int
}

type XPBoard struct {
	Entries     []XPEntry
	CurrentUser *XPEntry
}

type StreakEntry struct {
	Rank              int
	UserID            uuid.UUID
	DisplayName       string
	CurrentStreak     int
	LongestStreak     int
	StreakStartedDate time.Time
	LastChargeDate    time.Time
}

type StreakBoard struct {
	Entries     []StreakEntry
	CurrentUser *StreakEntry
}
