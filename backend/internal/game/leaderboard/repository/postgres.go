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

func (r *PostgreSQLRepository) GetTop(ctx context.Context) ([]leadermodel.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ROW_NUMBER() OVER (ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC), u.id, u.display_name,
		       COALESCE(p.xp, 0), COALESCE(p.xp / 100 + 1, 1)
		FROM users u
		LEFT JOIN pets p ON p.user_id = u.id
		ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []leadermodel.Entry
	for rows.Next() {
		var value leadermodel.Entry
		if err := rows.Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
func (r *PostgreSQLRepository) GetUserRank(ctx context.Context, userID uuid.UUID) (*leadermodel.Entry, error) {
	var value leadermodel.Entry
	err := r.db.QueryRowContext(ctx, `
		WITH ranked AS (
			SELECT ROW_NUMBER() OVER (ORDER BY COALESCE(p.xp, 0) DESC, u.id ASC) AS rank,
			       u.id, u.display_name, COALESCE(p.xp, 0) AS xp, COALESCE(p.xp / 100 + 1, 1) AS level
			FROM users u LEFT JOIN pets p ON p.user_id = u.id
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

func (r *PostgreSQLRepository) GetTopByDelta(ctx context.Context, since time.Time) ([]leadermodel.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH deltas AS (
			SELECT u.id, u.display_name,
			       COALESCE(p.xp, 0) AS xp,
			       COALESCE(p.xp / 100 + 1, 1) AS level,
			       COALESCE(SUM(xe.amount), 0) AS xp_delta
			FROM users u
			LEFT JOIN pets p ON p.user_id = u.id
			LEFT JOIN xp_events xe ON xe.user_id = u.id AND xe.occurred_at >= $1
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

	var result []leadermodel.Entry
	for rows.Next() {
		var value leadermodel.Entry
		if err := rows.Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level, &value.XPDelta); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) GetUserRankByDelta(ctx context.Context, userID uuid.UUID, since time.Time) (*leadermodel.Entry, error) {
	var value leadermodel.Entry
	err := r.db.QueryRowContext(ctx, `
		WITH deltas AS (
			SELECT u.id, u.display_name,
			       COALESCE(p.xp, 0) AS xp,
			       COALESCE(p.xp / 100 + 1, 1) AS level,
			       COALESCE(SUM(xe.amount), 0) AS xp_delta
			FROM users u
			LEFT JOIN pets p ON p.user_id = u.id
			LEFT JOIN xp_events xe ON xe.user_id = u.id AND xe.occurred_at >= $1
			GROUP BY u.id, u.display_name, p.xp
		),
		ranked AS (
			SELECT ROW_NUMBER() OVER (ORDER BY xp_delta DESC, id ASC) AS rank,
			       id, display_name, xp, level, xp_delta
			FROM deltas
		)
		SELECT rank, id, display_name, xp, level, xp_delta FROM ranked WHERE id = $2
	`, since, userID).Scan(&value.Rank, &value.UserID, &value.DisplayName, &value.XP, &value.Level, &value.XPDelta)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}
