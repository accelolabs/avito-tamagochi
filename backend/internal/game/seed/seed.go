package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	"github.com/google/uuid"
)

const (
	UserCount = 10
	xpLead    = 1_000_000
	xpStep    = 10_000
)

const seedEmailPattern = "leaderboard-seed-%02d@example.invalid"

type Result struct {
	Users int
}

func Seed(ctx context.Context, db *sql.DB, now time.Time) (Result, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE email ~ '^seed-[0-9]+@example\.com$'`); err != nil {
		return Result{}, fmt.Errorf("delete legacy seed users: %w", err)
	}

	now = now.UTC()
	today := clock.MoscowDate(now)
	maxXP, err := maximumXP(ctx, tx, now)
	if err != nil {
		return Result{}, err
	}
	maxStreak, err := maximumActiveStreak(ctx, tx, today)
	if err != nil {
		return Result{}, err
	}

	topXP := maxXP + xpLead + UserCount*xpStep
	topStreak := max(100, maxStreak+UserCount)
	for index := range UserCount {
		if err := upsertUser(ctx, tx, index, topXP-index*xpStep, topStreak-index, now, today); err != nil {
			return Result{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return Result{Users: UserCount}, nil
}

func maximumXP(ctx context.Context, tx *sql.Tx, now time.Time) (int, error) {
	var maximum int
	err := tx.QueryRowContext(ctx, `
		WITH totals AS (
			SELECT COALESCE(MAX(p.xp), 0) AS value
			FROM pets p JOIN users u ON u.id = p.user_id
			WHERE u.email NOT LIKE 'leaderboard-seed-%@example.invalid'
		), weekly AS (
			SELECT COALESCE(MAX(value), 0) AS value FROM (
				SELECT SUM(xe.amount) AS value
				FROM xp_events xe JOIN users u ON u.id = xe.user_id
				WHERE u.email NOT LIKE 'leaderboard-seed-%@example.invalid'
				  AND xe.occurred_at >= $1
				GROUP BY xe.user_id
			) values_by_user
		), monthly AS (
			SELECT COALESCE(MAX(value), 0) AS value FROM (
				SELECT SUM(xe.amount) AS value
				FROM xp_events xe JOIN users u ON u.id = xe.user_id
				WHERE u.email NOT LIKE 'leaderboard-seed-%@example.invalid'
				  AND xe.occurred_at >= $2
				GROUP BY xe.user_id
			) values_by_user
		)
		SELECT GREATEST(totals.value, weekly.value, monthly.value)
		FROM totals, weekly, monthly
	`, now.Add(-7*24*time.Hour), now.Add(-30*24*time.Hour)).Scan(&maximum)
	if err != nil {
		return 0, fmt.Errorf("read XP maximum: %w", err)
	}
	return maximum, nil
}

func maximumActiveStreak(ctx context.Context, tx *sql.Tx, today time.Time) (int, error) {
	var maximum int
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(p.charge_streak), 0)
		FROM pets p JOIN users u ON u.id = p.user_id
		WHERE u.email NOT LIKE 'leaderboard-seed-%@example.invalid'
		  AND p.last_streak_date >= $1::date - 1
	`, today.Format("2006-01-02")).Scan(&maximum)
	if err != nil {
		return 0, fmt.Errorf("read streak maximum: %w", err)
	}
	return maximum, nil
}

func upsertUser(ctx context.Context, tx *sql.Tx, index, xp, streak int, now, today time.Time) error {
	number := index + 1
	email := fmt.Sprintf(seedEmailPattern, number)
	displayName := fmt.Sprintf("Seed_%02d", number)
	userID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(email))
	petID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("pet:"+email))

	var existingID uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check seed user %d: %w", number, err)
	}
	if err == nil && existingID != userID {
		return fmt.Errorf("seed email %q belongs to a non-seed user", email)
	}

	createdAt := now.AddDate(0, 0, -streak)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, email, display_name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email, display_name = EXCLUDED.display_name
	`, userID, email, displayName, "$argon2id$seed$login-disabled", createdAt)
	if err != nil {
		return fmt.Errorf("upsert seed user %d: %w", number, err)
	}

	streakStarted := today.AddDate(0, 0, -(streak - 1))
	_, err = tx.ExecContext(ctx, `
		INSERT INTO pets (
			id, user_id, xp, energy_percent, energy_updated_at, last_charged_at,
			charge_streak, last_streak_date, longest_streak, streak_started_date,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, 100, $4, $4, $5, $6, $5, $7, $8, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET xp = EXCLUDED.xp,
		    energy_percent = EXCLUDED.energy_percent,
		    energy_updated_at = EXCLUDED.energy_updated_at,
		    last_charged_at = EXCLUDED.last_charged_at,
		    charge_streak = EXCLUDED.charge_streak,
		    last_streak_date = EXCLUDED.last_streak_date,
		    longest_streak = EXCLUDED.longest_streak,
		    streak_started_date = EXCLUDED.streak_started_date,
		    updated_at = EXCLUDED.updated_at
	`, petID, userID, xp, now, streak, today.Format("2006-01-02"), streakStarted.Format("2006-01-02"), createdAt)
	if err != nil {
		return fmt.Errorf("upsert seed pet %d: %w", number, err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM xp_events WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("replace seed XP events %d: %w", number, err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO xp_events (id, user_id, pet_id, source, source_key, amount, occurred_at, local_date)
		VALUES ($1, $2, $3, 'seed', 'seed:leaderboard', $4, $5, $6)
	`, uuid.NewSHA1(uuid.NameSpaceDNS, []byte("xp:"+email)), userID, petID, xp, now, today.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("insert seed XP event %d: %w", number, err)
	}
	return nil
}
