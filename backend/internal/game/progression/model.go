package progression

import (
	"time"

	"github.com/google/uuid"
)

type XPEvent struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	PetID      uuid.UUID
	Source     string
	SourceKey  string
	Amount     int
	OccurredAt time.Time
	LocalDate  time.Time
}
