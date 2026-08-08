package leaderboard

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	GetTop(context.Context, int) ([]Entry, error)
	GetUserRank(context.Context, uuid.UUID) (*Entry, error)
}
