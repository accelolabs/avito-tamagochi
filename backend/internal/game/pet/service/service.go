package service

import (
	"context"
	"database/sql"
	"errors"
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
	GetStreak(context.Context, uuid.UUID) (*petmodel.StreakStats, error)
	ChargePet(context.Context, uuid.UUID) (*petmodel.ChargeResult, error)
	Pet(context.Context, uuid.UUID) (*petmodel.PetActionResult, error)
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

	pet, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	reset := rules.IsDead(s.energy(pet, now)) && needsDeathReset(pet)
	if reset {
		if err := s.resetAfterDeath(ctx, tx, pet); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if reset {
		s.notifyDeath(userID)
	}
	return s.stats(pet, now), nil
}

func (s *service) GetStreak(ctx context.Context, userID uuid.UUID) (*petmodel.StreakStats, error) {
	pet, err := s.petRepo.GetByUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return &petmodel.StreakStats{NextDailyRewardXP: rules.DailyRewardXP(1)}, nil
	}
	if err != nil {
		return nil, err
	}
	today := clock.MoscowDate(s.clock.Now())
	current := rules.CurrentStreak(pet.ChargeStreak, pet.LastStreakDate, today)
	return &petmodel.StreakStats{
		CurrentStreak:     current,
		LongestStreak:     pet.LongestStreak,
		LastChargeDate:    pet.LastStreakDate,
		NextDailyRewardXP: rules.DailyRewardXP(current + 1),
	}, nil
}

func (s *service) getOrCreate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time) (*petmodel.Pet, error) {
	return s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, petmodel.NewPet(userID, now))
}

func (s *service) createXPEvents(ctx context.Context, tx *sql.Tx, events ...progressionmodel.XPEvent) error {
	for _, event := range events {
		if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) unlockLevels(ctx context.Context, tx *sql.Tx, userID uuid.UUID, oldXP, newXP int, now time.Time) (bool, error) {
	oldLevel := rules.LevelFromXP(oldXP)
	newLevel := rules.LevelFromXP(newXP)
	if s.rewardRepo == nil {
		return newLevel > oldLevel, nil
	}
	for level := oldLevel + 1; level <= newLevel; level++ {
		if err := s.rewardRepo.UnlockForLevel(ctx, tx, userID, level, rules.RewardTypeForLevel(level), now); err != nil {
			return false, err
		}
	}
	return newLevel > oldLevel, nil
}

func (s *service) resetAfterDeath(ctx context.Context, tx *sql.Tx, pet *petmodel.Pet) error {
	if err := s.petRepo.ResetAfterDeath(ctx, tx, pet.ID, pet.UserID); err != nil {
		return err
	}
	pet.XP = 0
	pet.ChargeStreak = 0
	pet.LastStreakDate = nil
	pet.StreakStartedDate = nil
	return nil
}

func (s *service) stats(pet *petmodel.Pet, now time.Time) *petmodel.Stats {
	level := rules.LevelFromXP(pet.XP)
	energy := s.energy(pet, now)
	return &petmodel.Stats{
		XP: pet.XP, Level: level, Stage: rules.StageFromLevel(level), Energy: energy,
		LastChargedAt: pet.LastChargedAt, IsDead: rules.IsDead(energy),
	}
}

func (s *service) energy(pet *petmodel.Pet, now time.Time) int {
	return rules.EnergyPercent(pet.EnergyPercent, pet.EnergyUpdatedAt, now)
}

func (s *service) notifyPetUpdated(userID uuid.UUID, rewardsUpdated bool) {
	if s.notify == nil {
		return
	}
	s.notify.NotifyUser(userID, "pet_updated")
	if rewardsUpdated {
		s.notify.NotifyUser(userID, "rewards_updated")
	}
}

func (s *service) notifyDeath(userID uuid.UUID) {
	s.notifyPetUpdated(userID, true)
}

func needsDeathReset(pet *petmodel.Pet) bool {
	return pet.XP > 0 || pet.ChargeStreak > 0 || pet.LastStreakDate != nil || pet.StreakStartedDate != nil
}
