package repository

import (
	"context"
	"database/sql"
	"time"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) GetTopByXP(ctx context.Context) ([]leadermodel.XPEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ROW_NUMBER() OVER (ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC),
		       u.id, u.display_name, COALESCE(p.xp, 0), COALESCE(p.xp / 100 + 1, 1)
		FROM users u
		LEFT JOIN pets p ON p.user_id = u.id
		ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]leadermodel.XPEntry, 0, 10)
	for rows.Next() {
		var value leadermodel.XPEntry
		if err := rows.Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) GetUserRankByXP(ctx context.Context, userID uuid.UUID) (*leadermodel.XPEntry, error) {
	var value leadermodel.XPEntry
	err := r.db.QueryRowContext(ctx, `
		WITH ranked AS (
			SELECT ROW_NUMBER() OVER (ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC) AS rank,
			       u.id, u.display_name, COALESCE(p.xp, 0) AS xp,
			       COALESCE(p.xp / 100 + 1, 1) AS level
			FROM users u
			LEFT JOIN pets p ON p.user_id = u.id
		)
		SELECT rank, id, display_name, xp, level FROM ranked WHERE id = $1
	`, userID).Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *PostgreSQLRepository) GetTopByXPDelta(ctx context.Context, since time.Time) ([]leadermodel.XPEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH deltas AS (
			SELECT u.id, u.display_name, COALESCE(p.xp, 0) AS xp,
			       COALESCE(p.xp / 100 + 1, 1) AS level, SUM(xe.amount) AS xp_delta
			FROM users u
			JOIN xp_events xe ON xe.user_id = u.id AND xe.occurred_at >= $1
			LEFT JOIN pets p ON p.user_id = u.id
			GROUP BY u.id, u.display_name, p.xp
		)
		SELECT ROW_NUMBER() OVER (ORDER BY xp_delta DESC, id ASC),
		       id, display_name, xp, level, xp_delta
		FROM deltas
		ORDER BY xp_delta DESC, id ASC
		LIMIT 10
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]leadermodel.XPEntry, 0, 10)
	for rows.Next() {
		var value leadermodel.XPEntry
		if err := rows.Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level, &value.XPDelta); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) GetUserRankByXPDelta(ctx context.Context, userID uuid.UUID, since time.Time) (*leadermodel.XPEntry, error) {
	var value leadermodel.XPEntry
	err := r.db.QueryRowContext(ctx, `
		WITH deltas AS (
			SELECT u.id, u.display_name, COALESCE(p.xp, 0) AS xp,
			       COALESCE(p.xp / 100 + 1, 1) AS level, SUM(xe.amount) AS xp_delta
			FROM users u
			JOIN xp_events xe ON xe.user_id = u.id AND xe.occurred_at >= $1
			LEFT JOIN pets p ON p.user_id = u.id
			GROUP BY u.id, u.display_name, p.xp
		), ranked AS (
			SELECT ROW_NUMBER() OVER (ORDER BY xp_delta DESC, id ASC) AS rank,
			       id, display_name, xp, level, xp_delta
			FROM deltas
		)
		SELECT rank, id, display_name, xp, level, xp_delta FROM ranked WHERE id = $2
	`, since, userID).Scan(
		&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level, &value.XPDelta,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *PostgreSQLRepository) GetTopByStreak(ctx context.Context, activeSince time.Time) ([]leadermodel.StreakEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH active AS (
			SELECT u.id, u.display_name, p.charge_streak, p.longest_streak,
			       p.streak_started_date, p.last_streak_date
			FROM users u
			JOIN pets p ON p.user_id = u.id
			WHERE p.charge_streak > 0 AND p.last_streak_date >= $1::date
		)
		SELECT ROW_NUMBER() OVER (
		           ORDER BY charge_streak DESC, streak_started_date ASC, id ASC
		       ), id, display_name, charge_streak, longest_streak,
		       streak_started_date, last_streak_date
		FROM active
		ORDER BY charge_streak DESC, streak_started_date ASC, id ASC
		LIMIT 10
	`, activeSince.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]leadermodel.StreakEntry, 0, 10)
	for rows.Next() {
		var value leadermodel.StreakEntry
		if err := scanStreakEntry(rows, &value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) GetUserRankByStreak(ctx context.Context, userID uuid.UUID, activeSince time.Time) (*leadermodel.StreakEntry, error) {
	var value leadermodel.StreakEntry
	err := scanStreakEntry(r.db.QueryRowContext(ctx, `
		WITH active AS (
			SELECT u.id, u.display_name, p.charge_streak, p.longest_streak,
			       p.streak_started_date, p.last_streak_date
			FROM users u
			JOIN pets p ON p.user_id = u.id
			WHERE p.charge_streak > 0 AND p.last_streak_date >= $1::date
		), ranked AS (
			SELECT ROW_NUMBER() OVER (
			           ORDER BY charge_streak DESC, streak_started_date ASC, id ASC
			       ) AS rank, id, display_name, charge_streak, longest_streak,
			       streak_started_date, last_streak_date
			FROM active
		)
		SELECT rank, id, display_name, charge_streak, longest_streak,
		       streak_started_date, last_streak_date
		FROM ranked WHERE id = $2
	`, activeSince.Format("2006-01-02"), userID), &value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

type rowScanner interface{ Scan(...any) error }

func scanStreakEntry(row rowScanner, value *leadermodel.StreakEntry) error {
	return row.Scan(
		&value.Rank, &value.UserID, &value.DisplayName, &value.CurrentStreak,
		&value.LongestStreak, &value.StreakStartedDate, &value.LastChargeDate,
	)
}
