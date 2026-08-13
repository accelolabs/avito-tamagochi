package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
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
	PetPet(context.Context, uuid.UUID) (*petmodel.PetActionResult, error)
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
	reset := rules.IsDead(s.energy(value, now)) && needsDeathReset(value)
	if reset {
		if err := s.resetAfterDeath(ctx, tx, value); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if reset {
		s.notifyDeath(userID)
	}
	return s.stats(value, now), nil
}

func (s *service) GetStreak(ctx context.Context, userID uuid.UUID) (*petmodel.StreakStats, error) {
	value, err := s.petRepo.GetByUser(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return &petmodel.StreakStats{NextDailyRewardXP: rules.DailyRewardXP(1)}, nil
	}
	if err != nil {
		return nil, err
	}
	today := clock.MoscowDate(s.clock.Now())
	current := rules.CurrentStreak(value.ChargeStreak, value.LastStreakDate, today)
	return &petmodel.StreakStats{
		CurrentStreak:     current,
		LongestStreak:     value.LongestStreak,
		LastChargeDate:    value.LastStreakDate,
		NextDailyRewardXP: rules.DailyRewardXP(current + 1),
	}, nil
}

func (s *service) ChargePet(ctx context.Context, userID uuid.UUID) (*petmodel.ChargeResult, error) {
	now := s.clock.Now().UTC()
	localDate := clock.MoscowDate(now)
	dailyRewardKey := fmt.Sprintf("daily_reward:%s", localDate.Format(time.DateOnly))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	value, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	currentEnergy := s.energy(value, now)
	if rules.IsDead(currentEnergy) && needsDeathReset(value) {
		if err := s.resetAfterDeath(ctx, tx, value); err != nil {
			return nil, err
		}
	}
	newEnergy := min(currentEnergy+rules.ChargeEnergyGain, 100)
	if newEnergy == currentEnergy {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &petmodel.ChargeResult{Pet: s.stats(value, now)}, nil
	}

	dailyRewardExists, err := s.xpRepo.HasSourceKey(ctx, tx, userID, dailyRewardKey)
	if err != nil {
		return nil, err
	}

	oldLevel := rules.LevelFromXP(value.XP)
	dailyReward := 0
	if !dailyRewardExists {
		s.advanceStreak(value, localDate)
		dailyReward = rules.DailyRewardXP(value.ChargeStreak)
	}
	totalAwarded := rules.ChargeXPAmount + dailyReward
	value.LastChargedAt = now
	value.EnergyPercent = newEnergy
	value.EnergyUpdatedAt = now
	value.XP += totalAwarded
	value.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *value); err != nil {
		return nil, err
	}
	if err := s.petRepo.ResetEnergyNotifications(ctx, tx, userID, newEnergy); err != nil {
		return nil, err
	}

	chargeEventID := uuid.New()
	events := []progressionmodel.XPEvent{
		{
			ID: chargeEventID, UserID: userID, PetID: value.ID, Source: "charge",
			SourceKey: "charge:" + chargeEventID.String(), Amount: rules.ChargeXPAmount, OccurredAt: now, LocalDate: localDate,
		},
	}
	if dailyReward > 0 {
		events = append(events, progressionmodel.XPEvent{
			ID: uuid.New(), UserID: userID, PetID: value.ID, Source: "daily_reward",
			SourceKey: dailyRewardKey, Amount: dailyReward, OccurredAt: now, LocalDate: localDate,
		})
	}
	for _, event := range events {
		if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
			return nil, err
		}
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
	return &petmodel.ChargeResult{
		Pet:            s.stats(value, now),
		BaseChargeXP:   rules.ChargeXPAmount,
		DailyRewardXP:  dailyReward,
		TotalXPAwarded: totalAwarded,
	}, nil
}

func (s *service) PetPet(ctx context.Context, userID uuid.UUID) (*petmodel.PetActionResult, error) {
	now := s.clock.Now().UTC()
	localDate := clock.MoscowDate(now)
	sourceKey := fmt.Sprintf("pet:%s", localDate.Format(time.DateOnly))
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	value, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	if rules.IsDead(s.energy(value, now)) {
		reset := needsDeathReset(value)
		if reset {
			if err := s.resetAfterDeath(ctx, tx, value); err != nil {
				return nil, err
			}
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		if reset {
			s.notifyDeath(userID)
		}
		return nil, gameerrors.ErrPetDead
	}

	exists, err := s.xpRepo.HasSourceKey(ctx, tx, userID, sourceKey)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &petmodel.PetActionResult{Pet: s.stats(value, now)}, nil
	}

	oldLevel := rules.LevelFromXP(value.XP)
	value.XP += rules.PetXPAmount
	value.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *value); err != nil {
		return nil, err
	}
	event := progressionmodel.XPEvent{
		ID: uuid.New(), UserID: userID, PetID: value.ID, Source: "pet",
		SourceKey: sourceKey, Amount: rules.PetXPAmount, OccurredAt: now, LocalDate: localDate,
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
	return &petmodel.PetActionResult{Pet: s.stats(value, now), XPAwarded: rules.PetXPAmount}, nil
}

func (s *service) getOrCreate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time) (*petmodel.Pet, error) {
	initial := petmodel.Pet{
		ID: uuid.New(), UserID: userID, EnergyPercent: 50, EnergyUpdatedAt: now, LastChargedAt: now.Add(-24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	return s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, initial)
}

func (s *service) resetAfterDeath(ctx context.Context, tx *sql.Tx, value *petmodel.Pet) error {
	if err := s.petRepo.ResetAfterDeath(ctx, tx, value.ID, value.UserID); err != nil {
		return err
	}
	value.XP = 0
	value.ChargeStreak = 0
	value.LastStreakDate = nil
	value.StreakStartedDate = nil
	return nil
}

func (s *service) advanceStreak(value *petmodel.Pet, today time.Time) {
	current := rules.CurrentStreak(value.ChargeStreak, value.LastStreakDate, today)
	if value.LastStreakDate == nil || !clock.MoscowDate(*value.LastStreakDate).Equal(today) {
		value.ChargeStreak = current + 1
	}
	if current == 0 {
		started := today
		value.StreakStartedDate = &started
	}
	if value.ChargeStreak > value.LongestStreak {
		value.LongestStreak = value.ChargeStreak
	}
	chargedOn := today
	value.LastStreakDate = &chargedOn
}

func (s *service) stats(value *petmodel.Pet, now time.Time) *petmodel.Stats {
	level := rules.LevelFromXP(value.XP)
	return &petmodel.Stats{
		XP: value.XP, Level: level, Stage: rules.StageFromLevel(level),
		Energy:        s.energy(value, now),
		LastChargedAt: value.LastChargedAt, IsDead: rules.IsDead(s.energy(value, now)),
	}
}

func (s *service) energy(value *petmodel.Pet, now time.Time) int {
	return rules.EnergyPercent(value.EnergyPercent, value.EnergyUpdatedAt, now)
}

func (s *service) notifyDeath(userID uuid.UUID) {
	if s.notify == nil {
		return
	}
	s.notify.NotifyUser(userID, "pet_updated")
	s.notify.NotifyUser(userID, "rewards_updated")
}

func needsDeathReset(value *petmodel.Pet) bool {
	return value.XP > 0 || value.ChargeStreak > 0 || value.LastStreakDate != nil || value.StreakStartedDate != nil
}
