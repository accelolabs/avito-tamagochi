package repository

import (
	"context"
	"database/sql"
	"time"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) UnlockForLevel(ctx context.Context, tx *sql.Tx, userID uuid.UUID, level int, rewardType rewardmodel.Type, unlockedAt time.Time) error {
	if level < 2 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO user_rewards (id, user_id, level, type, status, unlocked_at)
		VALUES ($1, $2, $3, $4, 'available', $5)
		ON CONFLICT (user_id, level) DO NOTHING
	`, uuid.New(), userID, level, rewardType, unlockedAt)
	return err
}

func (r *PostgreSQLRepository) GetUserRewards(ctx context.Context, userID uuid.UUID) ([]rewardmodel.UserReward, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ur.id, ur.user_id, ur.level, ur.type, ur.status, ur.unlocked_at, ur.used_at
		FROM user_rewards ur
		WHERE ur.user_id = $1
		ORDER BY ur.level
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []rewardmodel.UserReward
	for rows.Next() {
		var value rewardmodel.UserReward
		if err := rows.Scan(&value.ID, &value.UserID, &value.Level, &value.RewardType, &value.Status, &value.UnlockedAt, &value.UsedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) Use(ctx context.Context, tx *sql.Tx, userID, rewardID uuid.UUID, usedAt time.Time) (*rewardmodel.UserReward, error) {
	var value rewardmodel.UserReward
	err := tx.QueryRowContext(ctx, `
		UPDATE user_rewards ur
		SET status = 'used', used_at = $3
		WHERE ur.id = $1 AND ur.user_id = $2 AND ur.status = 'available'
		RETURNING ur.id, ur.user_id, ur.level, ur.type, ur.status, ur.unlocked_at, ur.used_at
	`, rewardID, userID, usedAt).Scan(&value.ID, &value.UserID, &value.Level, &value.RewardType, &value.Status, &value.UnlockedAt, &value.UsedAt)
	if err == nil {
		return &value, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	var status string
	err = tx.QueryRowContext(ctx, `SELECT status FROM user_rewards WHERE id = $1 AND user_id = $2`, rewardID, userID).Scan(&status)
	if err == sql.ErrNoRows {
		return nil, gameerrors.ErrRewardNotFound
	}
	if err != nil {
		return nil, err
	}
	return nil, gameerrors.ErrRewardAlreadyUsed
}
