package service

import (
	"context"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetRewards(context.Context, uuid.UUID) ([]rewardmodel.UserReward, error)
	UseReward(context.Context, uuid.UUID, uuid.UUID) (*rewardmodel.UserReward, error)
	UnlockForLevel(context.Context, uuid.UUID, int) error
}

type service struct{ repo rewardrepository.Repository }

func New(repo rewardrepository.Repository) Service { return &service{repo: repo} }
func (s *service) GetRewards(context.Context, uuid.UUID) ([]rewardmodel.UserReward, error) {
	return nil, gameerrors.ErrNotImplemented
}
func (s *service) UseReward(context.Context, uuid.UUID, uuid.UUID) (*rewardmodel.UserReward, error) {
	return nil, gameerrors.ErrNotImplemented
}
func (s *service) UnlockForLevel(context.Context, uuid.UUID, int) error {
	return gameerrors.ErrNotImplemented
}
