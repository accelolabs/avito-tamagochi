package leaderboard

import "github.com/google/uuid"

type Entry struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"userId"`
	DisplayName string    `json:"displayName"`
	XP          int       `json:"xp"`
	Level       int       `json:"level"`
}
