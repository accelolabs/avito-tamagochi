package repository

import (
	"context"
	"database/sql"

	progressionmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/model"
	"github.com/google/uuid"
)

type Repository interface {
	CreateXPEvent(context.Context, *sql.Tx, progressionmodel.XPEvent) error
	HasSourceKey(context.Context, *sql.Tx, uuid.UUID, string) (bool, error)
}
