package model

import "time"

type DailySummary struct {
	LocalDate       time.Time
	XPEarned        int
	CompletedTasks  int
	Charges         int
	Level           int
	CurrentXP       int
	Energy          int
	UnlockedRewards []string
}
