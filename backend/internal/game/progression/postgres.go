package progression

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

func (r *PostgreSQLRepository) CreateXPEvent(ctx context.Context, tx *sql.Tx, event XPEvent) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO xp_events (id, user_id, pet_id, source, source_key, amount, occurred_at, local_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.ID, event.UserID, event.PetID, event.Source, event.SourceKey, event.Amount, event.OccurredAt, event.LocalDate)
	return err
}

func (r *PostgreSQLRepository) HasSourceKey(ctx context.Context, tx *sql.Tx, userID uuid.UUID, sourceKey string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM xp_events WHERE user_id = $1 AND source_key = $2)
	`, userID, sourceKey).Scan(&exists)
	return exists, err
}
