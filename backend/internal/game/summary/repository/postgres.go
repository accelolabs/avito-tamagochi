package repository

import (
	"context"
	"database/sql"
	"time"

	summarymodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) GetToday(ctx context.Context, userID uuid.UUID, localDate, now time.Time) (*summarymodel.DailySummary, error) {
	value := &summarymodel.DailySummary{LocalDate: localDate, UnlockedRewards: make([]string, 0)}
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE((
		           SELECT SUM(xp.amount)
		           FROM xp_events xp
		           WHERE xp.user_id = u.id AND xp.local_date = $2::date
		       ), 0),
		       (
		           SELECT COUNT(*)
		           FROM task_progress tp
		           WHERE tp.user_id = u.id
		             AND tp.local_date = $2::date
		             AND tp.completed_at IS NOT NULL
		       ),
		       (
		           SELECT COUNT(*)
		           FROM xp_events xp
		           WHERE xp.user_id = u.id
		             AND xp.local_date = $2::date
		             AND xp.source = 'charge'
		       ),
		       COALESCE(p.xp, 0), COALESCE(p.xp / 100 + 1, 1),
		       CASE WHEN p.id IS NULL THEN 0 ELSE GREATEST(0, LEAST(100,
		           p.energy_percent - FLOOR(EXTRACT(EPOCH FROM ($3 - p.energy_updated_at))
		           / EXTRACT(EPOCH FROM INTERVAL '48 hours') * 100)::int
		       )) END
		FROM users u
		LEFT JOIN pets p ON p.user_id = u.id
		WHERE u.id = $1
	`, userID, localDate.Format("2006-01-02"), now).Scan(&value.XPEarned, &value.CompletedTasks, &value.Charges, &value.CurrentXP, &value.Level, &value.Energy)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT ur.type FROM user_rewards ur
		WHERE ur.user_id = $1 AND ur.unlocked_at >= $2 AND ur.unlocked_at < $2 + INTERVAL '1 day'
		ORDER BY ur.unlocked_at`, userID, localDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var reward string
		if err := rows.Scan(&reward); err != nil {
			return nil, err
		}
		value.UnlockedRewards = append(value.UnlockedRewards, reward)
	}
	return value, rows.Err()
}
