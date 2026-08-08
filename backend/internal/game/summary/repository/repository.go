package repository

import (
	"context"
	"time"

	summarymodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetToday(context.Context, uuid.UUID, time.Time) (*summarymodel.DailySummary, error)
}
