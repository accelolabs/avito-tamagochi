package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Recipient string
	Subject   string
	TextBody  string
	HTMLBody  string
}

type Participant struct {
	UserID          uuid.UUID
	Email           string
	EnergyPercent   int
	EnergyUpdatedAt time.Time
}

type BatchResult struct {
	Participants int
	Sent         int
	Skipped      int
	Failed       int
}
