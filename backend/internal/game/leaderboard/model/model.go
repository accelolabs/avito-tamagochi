package model

import "github.com/google/uuid"

type Entry struct {
	Rank        int
	UserID      uuid.UUID
	DisplayName string
	XP          int
	Level       int
	XPDelta     int
}
