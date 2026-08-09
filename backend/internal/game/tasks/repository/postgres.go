package repository

import (
	"context"
	"database/sql"
	"time"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) GetTodayProgress(ctx context.Context, userID uuid.UUID, localDate time.Time) ([]taskmodel.Progress, error) {
	date := localDate.Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx, `
		SELECT $1, $2, r.task_type, d.title, d.xp_reward,
		       COALESCE(p.progress, 0), d.required_count, p.completed_at
		FROM task_rotation r
		JOIN task_definitions d ON d.type = r.task_type
		LEFT JOIN task_progress p ON p.task_type = r.task_type AND p.user_id = $1 AND p.local_date = $2
		WHERE r.cycle_day = ((($2::date - DATE '1970-01-01') % 3) + 1)
		ORDER BY r.task_type`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]taskmodel.Progress, 0)
	for rows.Next() {
		value, err := scanProgress(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PostgreSQLRepository) GetProgressForUpdate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, localDate time.Time, taskType taskmodel.Type) (*taskmodel.Progress, error) {
	date := localDate.Format("2006-01-02")
	_, err := tx.ExecContext(ctx, `
		INSERT INTO task_progress (user_id, local_date, task_type)
		SELECT $1, $2, r.task_type
		FROM task_rotation r
		WHERE r.cycle_day = ((($2::date - DATE '1970-01-01') % 3) + 1) AND r.task_type = $3
		ON CONFLICT (user_id, local_date, task_type) DO NOTHING
	`, userID, date, taskType)
	if err != nil {
		return nil, err
	}

	return scanProgress(tx.QueryRowContext(ctx, taskQuery+` WHERE p.user_id = $1 AND p.local_date = $2::date AND p.task_type = $3 FOR UPDATE`, userID, date, taskType))
}

func (r *PostgreSQLRepository) SaveProgress(ctx context.Context, tx *sql.Tx, value taskmodel.Progress) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE task_progress
		SET progress = $1, completed_at = $2
		WHERE user_id = $3 AND local_date = $4 AND task_type = $5
	`, value.Progress, value.CompletedAt, value.UserID, value.LocalDate, value.TaskType)
	return err
}

const taskQuery = `
	SELECT p.user_id, p.local_date, p.task_type, d.title, d.xp_reward,
	       p.progress, d.required_count, p.completed_at
	FROM task_progress p
	JOIN task_definitions d ON d.type = p.task_type
	JOIN task_rotation r ON r.task_type = p.task_type
`

type scanner interface{ Scan(...any) error }

func scanProgress(row scanner) (*taskmodel.Progress, error) {
	var value taskmodel.Progress
	var completedAt *time.Time
	if err := row.Scan(&value.UserID, &value.LocalDate, &value.TaskType, &value.Title, &value.XPReward, &value.Progress, &value.RequiredCount, &completedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, gameerrors.ErrTaskNotAvailable
		}
		return nil, err
	}
	value.CompletedAt = completedAt
	value.Completed = completedAt != nil
	return &value, nil
}
