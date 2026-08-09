package repository

import (
	"context"
	"database/sql"

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
