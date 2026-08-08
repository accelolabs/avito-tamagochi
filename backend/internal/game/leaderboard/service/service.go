package service

import (
	"context"

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
func (s *service) GetTop(ctx context.Context, limit int) ([]leadermodel.Entry, error) {
	if limit != 10 {
		limit = 10
	}
	return s.repo.GetTop(ctx, limit)
}
func (s *service) GetUserRank(ctx context.Context, userID uuid.UUID) (*leadermodel.Entry, error) {
	return s.repo.GetUserRank(ctx, userID)
}
