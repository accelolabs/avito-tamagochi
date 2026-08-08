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
	value := &summarymodel.DailySummary{LocalDate: localDate}
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(xp.amount), 0),
		       COUNT(*) FILTER (WHERE xp.source = 'task'),
		       COUNT(*) FILTER (WHERE xp.source = 'charge'),
		       COALESCE(p.xp, 0), COALESCE(p.xp / 100 + 1, 1),
		       CASE WHEN p.id IS NULL THEN 0 ELSE 100 - LEAST(100, GREATEST(0, FLOOR(EXTRACT(EPOCH FROM ($3 - p.last_charged_at)) / 1728)))::int END
		FROM users u
		LEFT JOIN pets p ON p.user_id = u.id
		LEFT JOIN xp_events xp ON xp.user_id = p.user_id AND xp.local_date = $2
		WHERE u.id = $1
		GROUP BY p.id, p.xp, p.last_charged_at`, userID, localDate.Format("2006-01-02"), now).Scan(&value.XPEarned, &value.CompletedTasks, &value.Charges, &value.CurrentXP, &value.Level, &value.Energy)
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
