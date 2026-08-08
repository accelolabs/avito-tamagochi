package repository

import (
	"context"
	"database/sql"

	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	"github.com/google/uuid"
)

type Repository interface {
	UnlockForLevel(context.Context, *sql.Tx, uuid.UUID, int) error
	GetUserRewards(context.Context, uuid.UUID) ([]rewardmodel.UserReward, error)
	Use(context.Context, *sql.Tx, uuid.UUID, uuid.UUID) (*rewardmodel.UserReward, error)
}
