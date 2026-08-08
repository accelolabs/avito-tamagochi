package service

import (
	"context"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	leaderrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetTop(context.Context, int) ([]leadermodel.Entry, error)
	GetUserRank(context.Context, uuid.UUID) (*leadermodel.Entry, error)
}

type service struct{ repo leaderrepository.Repository }

func New(repo leaderrepository.Repository) Service { return &service{repo: repo} }
func (s *service) GetTop(context.Context, int) ([]leadermodel.Entry, error) {
	return nil, gameerrors.ErrNotImplemented
}
func (s *service) GetUserRank(context.Context, uuid.UUID) (*leadermodel.Entry, error) {
	return nil, gameerrors.ErrNotImplemented
}
