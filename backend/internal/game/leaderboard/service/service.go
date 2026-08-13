package service

import (
	"context"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	leaderrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetTop(context.Context) ([]leadermodel.Entry, error)
	GetUserRank(context.Context, uuid.UUID) (*leadermodel.Entry, error)
	GetTopByPeriod(ctx context.Context, period string) ([]leadermodel.Entry, error)
	GetUserRankByPeriod(ctx context.Context, userID uuid.UUID, period string) (*leadermodel.Entry, error)
}
type service struct {
	repo  leaderrepository.Repository
	clock clock.Clock
}

func New(repo leaderrepository.Repository) Service {
	return &service{repo: repo, clock: clock.RealClock{}}
}

func NewWithClock(repo leaderrepository.Repository, c clock.Clock) Service {
	if c == nil {
		c = clock.RealClock{}
	}
	return &service{repo: repo, clock: c}
}

func (s *service) GetTop(ctx context.Context) ([]leadermodel.Entry, error) {
	return s.repo.GetTop(ctx)
}

func (s *service) GetUserRank(ctx context.Context, userID uuid.UUID) (*leadermodel.Entry, error) {
	return s.repo.GetUserRank(ctx, userID)
}

func (s *service) GetTopByPeriod(ctx context.Context, period string) ([]leadermodel.Entry, error) {
	since, ok := s.sinceFromPeriod(period)
	if !ok {
		return s.repo.GetTop(ctx)
	}
	return s.repo.GetTopByDelta(ctx, since)
}

func (s *service) GetUserRankByPeriod(ctx context.Context, userID uuid.UUID, period string) (*leadermodel.Entry, error) {
	since, ok := s.sinceFromPeriod(period)
	if !ok {
		return s.repo.GetUserRank(ctx, userID)
	}
	return s.repo.GetUserRankByDelta(ctx, userID, since)
}

func (s *service) sinceFromPeriod(period string) (time.Time, bool) {
	now := s.clock.Now().UTC()
	today := clock.MoscowDate(now)
	switch period {
	case "week":
		return today.AddDate(0, 0, -7), true
	case "month":
		return today.AddDate(0, 0, -30), true
	default:
		return time.Time{}, false
	}
}
