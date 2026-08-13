package repository

import (
	"context"
	"time"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetTopByXP(context.Context) ([]leadermodel.XPEntry, error)
	GetUserRankByXP(context.Context, uuid.UUID) (*leadermodel.XPEntry, error)
	GetTopByXPDelta(context.Context, time.Time) ([]leadermodel.XPEntry, error)
	GetUserRankByXPDelta(context.Context, uuid.UUID, time.Time) (*leadermodel.XPEntry, error)
	GetTopByStreak(context.Context, time.Time) ([]leadermodel.StreakEntry, error)
	GetUserRankByStreak(context.Context, uuid.UUID, time.Time) (*leadermodel.StreakEntry, error)
}
