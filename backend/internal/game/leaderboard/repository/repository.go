package repository

import (
	"context"
	"time"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetTop(context.Context) ([]leadermodel.Entry, error)
	GetUserRank(context.Context, uuid.UUID) (*leadermodel.Entry, error)
	GetTopByDelta(ctx context.Context, since time.Time) ([]leadermodel.Entry, error)
	GetUserRankByDelta(ctx context.Context, userID uuid.UUID, since time.Time) (*leadermodel.Entry, error)
}
