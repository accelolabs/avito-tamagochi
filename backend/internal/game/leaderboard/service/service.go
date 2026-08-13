package service

import (
	"context"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	leaderrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/repository"
	"github.com/google/uuid"
)

const (
	weekWindow  = 7 * 24 * time.Hour
	monthWindow = 30 * 24 * time.Hour
)

type Service interface {
	GetAll(context.Context, uuid.UUID) (*leadermodel.XPBoard, error)
	GetWeekly(context.Context, uuid.UUID) (*leadermodel.XPBoard, error)
	GetMonthly(context.Context, uuid.UUID) (*leadermodel.XPBoard, error)
	GetStreak(context.Context, uuid.UUID) (*leadermodel.StreakBoard, error)
}

type service struct {
	repo  leaderrepository.Repository
	clock clock.Clock
}

func New(repo leaderrepository.Repository) Service {
	return NewWithClock(repo, clock.RealClock{})
}

func NewWithClock(repo leaderrepository.Repository, currentClock clock.Clock) Service {
	if currentClock == nil {
		currentClock = clock.RealClock{}
	}
	return &service{repo: repo, clock: currentClock}
}

func (s *service) GetAll(ctx context.Context, userID uuid.UUID) (*leadermodel.XPBoard, error) {
	entries, err := s.repo.GetTopByXP(ctx)
	if err != nil {
		return nil, err
	}
	currentUser, err := s.repo.GetUserRankByXP(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &leadermodel.XPBoard{Entries: entries, CurrentUser: currentUser}, nil
}

func (s *service) GetWeekly(ctx context.Context, userID uuid.UUID) (*leadermodel.XPBoard, error) {
	return s.getDelta(ctx, userID, s.clock.Now().UTC().Add(-weekWindow))
}

func (s *service) GetMonthly(ctx context.Context, userID uuid.UUID) (*leadermodel.XPBoard, error) {
	return s.getDelta(ctx, userID, s.clock.Now().UTC().Add(-monthWindow))
}

func (s *service) GetStreak(ctx context.Context, userID uuid.UUID) (*leadermodel.StreakBoard, error) {
	today := clock.MoscowDate(s.clock.Now())
	activeSince := today.AddDate(0, 0, -1)
	entries, err := s.repo.GetTopByStreak(ctx, activeSince)
	if err != nil {
		return nil, err
	}
	currentUser, err := s.repo.GetUserRankByStreak(ctx, userID, activeSince)
	if err != nil {
		return nil, err
	}
	return &leadermodel.StreakBoard{Entries: entries, CurrentUser: currentUser}, nil
}

func (s *service) getDelta(ctx context.Context, userID uuid.UUID, since time.Time) (*leadermodel.XPBoard, error) {
	entries, err := s.repo.GetTopByXPDelta(ctx, since)
	if err != nil {
		return nil, err
	}
	currentUser, err := s.repo.GetUserRankByXPDelta(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	return &leadermodel.XPBoard{Entries: entries, CurrentUser: currentUser}, nil
}
