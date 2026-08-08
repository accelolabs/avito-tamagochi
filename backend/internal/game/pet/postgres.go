package pet

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

func (r *PostgreSQLRepository) GetByUser(ctx context.Context, userID uuid.UUID) (*Pet, error) {
	return scanPet(r.db.QueryRowContext(ctx, petSelect+" WHERE user_id = $1", userID))
}

func (r *PostgreSQLRepository) GetOrCreateForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, initial Pet) (*Pet, error) {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO pets (id, user_id, xp, last_charged_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO NOTHING
	`, initial.ID, initial.UserID, initial.XP, initial.LastChargedAt, initial.CreatedAt, initial.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return scanPet(tx.QueryRowContext(ctx, petSelect+" WHERE user_id = $1 FOR UPDATE", userID))
}

func (r *PostgreSQLRepository) Update(ctx context.Context, tx *sql.Tx, value Pet) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE pets
		SET xp = $1, last_charged_at = $2, updated_at = $3
		WHERE id = $4
	`, value.XP, value.LastChargedAt, value.UpdatedAt, value.ID)
	return err
}

const petSelect = `SELECT id, user_id, xp, last_charged_at, created_at, updated_at FROM pets`

type rowScanner interface {
	Scan(...any) error
}

func scanPet(row rowScanner) (*Pet, error) {
	var value Pet
	if err := row.Scan(&value.ID, &value.UserID, &value.XP, &value.LastChargedAt, &value.CreatedAt, &value.UpdatedAt); err != nil {
		return nil, err
	}
	return &value, nil
}
