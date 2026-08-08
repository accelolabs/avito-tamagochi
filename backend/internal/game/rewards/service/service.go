package service

import (
	"context"
	"database/sql"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/notifier"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetRewards(context.Context, uuid.UUID) ([]rewardmodel.UserReward, error)
	UseReward(context.Context, uuid.UUID, uuid.UUID) (*rewardmodel.UserReward, error)
}

type service struct {
	db     *sql.DB
	repo   rewardrepository.Repository
	notify notifier.Notifier
}

func New(db *sql.DB, repo rewardrepository.Repository, notify notifier.Notifier) Service {
	return &service{db: db, repo: repo, notify: notify}
}
func (s *service) GetRewards(ctx context.Context, userID uuid.UUID) ([]rewardmodel.UserReward, error) {
	return s.repo.GetUserRewards(ctx, userID)
}
func (s *service) UseReward(ctx context.Context, userID, rewardID uuid.UUID) (*rewardmodel.UserReward, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	value, err := s.repo.Use(ctx, tx, userID, rewardID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if s.notify != nil {
		s.notify.NotifyUser(userID, "rewards_updated")
	}
	return value, nil
}
