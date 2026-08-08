package rewards

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	UnlockForLevel(context.Context, *sql.Tx, uuid.UUID, int) error
	GetUserRewards(context.Context, uuid.UUID) ([]UserReward, error)
	Use(context.Context, *sql.Tx, uuid.UUID, uuid.UUID) (*UserReward, error)
}
