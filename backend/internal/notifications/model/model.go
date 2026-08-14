package model

import (
	"context"
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

type DeliveredThresholds map[int]bool

type ParticipantHandler func(context.Context, Participant, DeliveredThresholds) (*int, error)

type BatchResult struct {
	Participants int
	Sent         int
	Skipped      int
	Failed       int
}
