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

func (r *PostgreSQLRepository) GetToday(ctx context.Context, userID uuid.UUID, localDate time.Time) (*summarymodel.DailySummary, error) {
	value := &summarymodel.DailySummary{LocalDate: localDate}
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(xp.amount), 0),
		       COUNT(*) FILTER (WHERE xp.source = 'task'),
		       COUNT(*) FILTER (WHERE xp.source = 'charge'),
		       p.xp, p.xp / 100 + 1,
		       100 - LEAST(100, GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - p.last_charged_at)) / 1728)))::int
		FROM pets p
		LEFT JOIN xp_events xp ON xp.user_id = p.user_id AND xp.local_date = $2
		WHERE p.user_id = $1
		GROUP BY p.xp, p.last_charged_at`, userID, localDate).Scan(&value.XPEarned, &value.CompletedTasks, &value.Charges, &value.CurrentXP, &value.Level, &value.Energy)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT rd.type FROM user_rewards ur JOIN reward_definitions rd ON rd.id = ur.reward_id
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
