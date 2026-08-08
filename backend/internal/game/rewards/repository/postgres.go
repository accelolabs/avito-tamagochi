package repository

import (
	"context"
	"database/sql"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) UnlockForLevel(ctx context.Context, tx *sql.Tx, userID uuid.UUID, level int) error {
	var rewardID uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT id FROM reward_definitions WHERE level = $1`, level).Scan(&rewardID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_rewards (id, user_id, reward_id, status, unlocked_at)
		VALUES ($1, $2, $3, 'available', NOW())
		ON CONFLICT (user_id, reward_id) DO NOTHING
	`, uuid.New(), userID, rewardID)
	return err
}

func (r *PostgreSQLRepository) GetUserRewards(ctx context.Context, userID uuid.UUID) ([]rewardmodel.UserReward, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ur.id, ur.user_id, rd.type, ur.status, ur.unlocked_at, ur.used_at
		FROM user_rewards ur
		JOIN reward_definitions rd ON rd.id = ur.reward_id
		WHERE ur.user_id = $1
		ORDER BY rd.level
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []rewardmodel.UserReward
	for rows.Next() {
		var value rewardmodel.UserReward
		if err := rows.Scan(&value.ID, &value.UserID, &value.RewardType, &value.Status, &value.UnlockedAt, &value.UsedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgreSQLRepository) Use(ctx context.Context, tx *sql.Tx, userID, rewardID uuid.UUID) (*rewardmodel.UserReward, error) {
	var value rewardmodel.UserReward
	err := tx.QueryRowContext(ctx, `
		UPDATE user_rewards ur
		SET status = 'used', used_at = NOW()
		FROM reward_definitions rd
		WHERE ur.reward_id = rd.id AND ur.id = $1 AND ur.user_id = $2 AND ur.status = 'available'
		RETURNING ur.id, ur.user_id, rd.type, ur.status, ur.unlocked_at, ur.used_at
	`, rewardID, userID).Scan(&value.ID, &value.UserID, &value.RewardType, &value.Status, &value.UnlockedAt, &value.UsedAt)
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
