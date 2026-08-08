package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/notifier"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	progressionmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/model"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetPet(context.Context, uuid.UUID) (*petmodel.Stats, error)
	ChargePet(context.Context, uuid.UUID) (*petmodel.Stats, error)
}

type service struct {
	db         *sql.DB
	petRepo    petrepository.Repository
	xpRepo     progressionrepository.Repository
	clock      clock.Clock
	rewardRepo rewardrepository.Repository
	notify     notifier.Notifier
}

func New(db *sql.DB, petRepo petrepository.Repository, xpRepo progressionrepository.Repository, rewardRepo rewardrepository.Repository, now clock.Clock, notify notifier.Notifier) Service {
	if now == nil {
		now = clock.RealClock{}
	}
	return &service{db: db, petRepo: petRepo, xpRepo: xpRepo, rewardRepo: rewardRepo, clock: now, notify: notify}
}

func (s *service) GetPet(ctx context.Context, userID uuid.UUID) (*petmodel.Stats, error) {
	now := s.clock.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	value, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.stats(value, now), nil
}

func (s *service) ChargePet(ctx context.Context, userID uuid.UUID) (*petmodel.Stats, error) {
	now := s.clock.Now().UTC()
	localDate := clock.MoscowDate(now)
	sourceKey := fmt.Sprintf("charge:%s", localDate.Format("2006-01-02"))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	value, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	exists, err := s.xpRepo.HasSourceKey(ctx, tx, userID, sourceKey)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return s.stats(value, now), nil
	}
	oldLevel := rules.LevelFromXP(value.XP)
	value.LastChargedAt = now
	value.XP += rules.ChargeXPAmount
	value.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *value); err != nil {
		return nil, err
	}
	event := progressionmodel.XPEvent{
		ID: uuid.New(), UserID: userID, PetID: value.ID, Source: "charge", SourceKey: sourceKey,
		Amount: rules.ChargeXPAmount, OccurredAt: now, LocalDate: localDate,
	}
	if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
		return nil, err
	}
	newLevel := rules.LevelFromXP(value.XP)
	for level := oldLevel + 1; level <= newLevel; level++ {
		if s.rewardRepo != nil {
			if err := s.rewardRepo.UnlockForLevel(ctx, tx, userID, level, rules.RewardTypeForLevel(level), now); err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if s.notify != nil {
		s.notify.NotifyUser(userID, "pet_updated")
		if newLevel > oldLevel {
			s.notify.NotifyUser(userID, "rewards_updated")
		}
	}
	return s.stats(value, now), nil
}

func (s *service) getOrCreate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time) (*petmodel.Pet, error) {
	initial := petmodel.Pet{ID: uuid.New(), UserID: userID, LastChargedAt: now.Add(-24 * time.Hour), CreatedAt: now, UpdatedAt: now}
	return s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, initial)
}

func (s *service) stats(value *petmodel.Pet, now time.Time) *petmodel.Stats {
	return &petmodel.Stats{XP: value.XP, Level: rules.LevelFromXP(value.XP), Energy: rules.EnergyPercent(value.LastChargedAt, now), LastChargedAt: value.LastChargedAt}
}
