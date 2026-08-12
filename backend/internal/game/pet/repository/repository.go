package repository

import (
	"context"
	"database/sql"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetByUser(context.Context, uuid.UUID) (*model.Pet, error)
	GetOrCreateForUpdate(context.Context, *sql.Tx, uuid.UUID, model.Pet) (*model.Pet, error)
	Update(context.Context, *sql.Tx, model.Pet) error
	ResetProgress(ctx context.Context, tx *sql.Tx, petID, userID uuid.UUID) error
}

