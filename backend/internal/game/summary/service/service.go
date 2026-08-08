package service

import (
	"context"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	summarymodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/model"
	summaryrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetToday(context.Context, uuid.UUID) (*summarymodel.DailySummary, error)
}

type service struct {
	repo summaryrepository.Repository
	now  clock.Clock
}

func New(repo summaryrepository.Repository, now clock.Clock) Service {
	if now == nil {
		now = clock.RealClock{}
	}
	return &service{repo: repo, now: now}
}
func (s *service) GetToday(ctx context.Context, userID uuid.UUID) (*summarymodel.DailySummary, error) {
	now := s.now.Now().UTC()
	return s.repo.GetToday(ctx, userID, clock.MoscowDate(now), now)
}
