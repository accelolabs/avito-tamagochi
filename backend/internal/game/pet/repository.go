package pet

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	GetByUser(context.Context, uuid.UUID) (*Pet, error)
	GetOrCreateForUpdate(context.Context, *sql.Tx, uuid.UUID, Pet) (*Pet, error)
	Update(context.Context, *sql.Tx, Pet) error
}
