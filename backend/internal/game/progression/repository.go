package progression

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	CreateXPEvent(context.Context, *sql.Tx, XPEvent) error
	HasSourceKey(context.Context, *sql.Tx, uuid.UUID, string) (bool, error)
}
