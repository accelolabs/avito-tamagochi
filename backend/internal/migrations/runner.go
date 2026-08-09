package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// Apply applies all pending PostgreSQL migrations from directory.
func Apply(ctx context.Context, db *sql.DB, directory string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.UpContext(ctx, db, directory)
}
