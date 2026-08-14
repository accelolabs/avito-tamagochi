package repository

import (
	"context"
	"database/sql"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) GetByUser(ctx context.Context, userID uuid.UUID) (*model.Pet, error) {
	return scanPet(r.db.QueryRowContext(ctx, petSelect+" WHERE user_id = $1", userID))
}

func (r *PostgreSQLRepository) GetOrCreateForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, initial model.Pet) (*model.Pet, error) {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO pets (
			id, user_id, xp, energy_percent, energy_updated_at, last_charged_at,
			charge_streak, longest_streak, last_streak_date, streak_started_date,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (user_id) DO NOTHING
	`, initial.ID, initial.UserID, initial.XP, initial.EnergyPercent, initial.EnergyUpdatedAt,
		initial.LastChargedAt, initial.ChargeStreak, initial.LongestStreak,
		initial.LastStreakDate, initial.StreakStartedDate, initial.CreatedAt, initial.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return scanPet(tx.QueryRowContext(ctx, petSelect+" WHERE user_id = $1 FOR UPDATE", userID))
}

func (r *PostgreSQLRepository) Update(ctx context.Context, tx *sql.Tx, value model.Pet) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE pets
		SET xp = $1,
		    energy_percent = $2,
		    energy_updated_at = $3,
		    last_charged_at = $4,
		    charge_streak = $5,
		    longest_streak = $6,
		    last_streak_date = $7,
		    streak_started_date = $8,
		    updated_at = $9
		WHERE id = $10
	`, value.XP, value.EnergyPercent, value.EnergyUpdatedAt, value.LastChargedAt,
		value.ChargeStreak, value.LongestStreak, value.LastStreakDate,
		value.StreakStartedDate, value.UpdatedAt, value.ID)
	return err
}

func (r *PostgreSQLRepository) ResetEnergyNotifications(ctx context.Context, tx *sql.Tx, userID uuid.UUID, energy int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM energy_notification_deliveries
		WHERE user_id = $1 AND threshold < $2
	`, userID, energy)
	return err
}

func (r *PostgreSQLRepository) ResetAfterDeath(ctx context.Context, tx *sql.Tx, petID, userID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE pets
		SET xp = 0, charge_streak = 0, last_streak_date = NULL, streak_started_date = NULL
		WHERE id = $1
	`, petID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		DELETE FROM user_rewards WHERE user_id = $1 AND status = 'available'
	`, userID)
	return err
}

const petSelect = `
	SELECT id, user_id, xp, energy_percent, energy_updated_at, last_charged_at, charge_streak, longest_streak,
	       last_streak_date, streak_started_date, created_at, updated_at
	FROM pets`

type rowScanner interface{ Scan(...any) error }

func scanPet(row rowScanner) (*model.Pet, error) {
	var value model.Pet
	if err := row.Scan(
		&value.ID, &value.UserID, &value.XP, &value.EnergyPercent, &value.EnergyUpdatedAt, &value.LastChargedAt,
		&value.ChargeStreak, &value.LongestStreak, &value.LastStreakDate,
		&value.StreakStartedDate, &value.CreatedAt, &value.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &value, nil
}
