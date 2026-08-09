package repository

import (
	"context"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetTop(context.Context) ([]leadermodel.Entry, error)
	GetUserRank(context.Context, uuid.UUID) (*leadermodel.Entry, error)
}
