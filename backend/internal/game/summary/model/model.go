package model

import "time"

type DailySummary struct {
	LocalDate       time.Time
	XPEarned        int
	CompletedTasks  int
	Charges         int
	Level           int
	UnlockedRewards []string
}
