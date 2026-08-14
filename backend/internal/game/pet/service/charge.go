package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	progressionmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/model"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
	"github.com/google/uuid"
)

func (s *service) ChargePet(ctx context.Context, userID uuid.UUID) (*petmodel.ChargeResult, error) {
	now := s.clock.Now().UTC()
	localDate := clock.MoscowDate(now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	pet, err := s.getOrCreate(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	currentEnergy := s.energy(pet, now)
	if rules.IsDead(currentEnergy) && needsDeathReset(pet) {
		if err := s.resetAfterDeath(ctx, tx, pet); err != nil {
			return nil, err
		}
	}
	newEnergy := min(currentEnergy+rules.ChargeEnergyGain, 100)
	if newEnergy == currentEnergy {
		return s.unchangedCharge(tx, pet, now)
	}

	dailyReward, err := s.dailyChargeReward(ctx, tx, userID, pet, localDate)
	if err != nil {
		return nil, err
	}
	oldXP := pet.XP
	totalAwarded := rules.ChargeXPAmount + dailyReward
	applyCharge(pet, newEnergy, totalAwarded, now)
	if err := s.persistCharge(ctx, tx, pet); err != nil {
		return nil, err
	}
	if err := s.createXPEvents(ctx, tx, chargeEvents(userID, pet.ID, dailyReward, localDate, now)...); err != nil {
		return nil, err
	}
	rewardsUpdated, err := s.unlockLevels(ctx, tx, userID, oldXP, pet.XP, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.notifyPetUpdated(userID, rewardsUpdated)
	return &petmodel.ChargeResult{
		Pet: s.stats(pet, now), BaseChargeXP: rules.ChargeXPAmount,
		DailyRewardXP: dailyReward, TotalXPAwarded: totalAwarded,
	}, nil
}

func (s *service) unchangedCharge(tx *sql.Tx, pet *petmodel.Pet, now time.Time) (*petmodel.ChargeResult, error) {
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &petmodel.ChargeResult{Pet: s.stats(pet, now)}, nil
}

func (s *service) dailyChargeReward(ctx context.Context, tx *sql.Tx, userID uuid.UUID, pet *petmodel.Pet, localDate time.Time) (int, error) {
	sourceKey := dailyRewardKey(localDate)
	exists, err := s.xpRepo.HasSourceKey(ctx, tx, userID, sourceKey)
	if err != nil || exists {
		return 0, err
	}
	s.advanceStreak(pet, localDate)
	return rules.DailyRewardXP(pet.ChargeStreak), nil
}

func (s *service) persistCharge(ctx context.Context, tx *sql.Tx, pet *petmodel.Pet) error {
	if err := s.petRepo.Update(ctx, tx, *pet); err != nil {
		return err
	}
	return s.petRepo.ResetEnergyNotifications(ctx, tx, pet.UserID, pet.EnergyPercent)
}

func (s *service) advanceStreak(pet *petmodel.Pet, today time.Time) {
	current := rules.CurrentStreak(pet.ChargeStreak, pet.LastStreakDate, today)
	if pet.LastStreakDate == nil || !clock.MoscowDate(*pet.LastStreakDate).Equal(today) {
		pet.ChargeStreak = current + 1
	}
	if current == 0 {
		started := today
		pet.StreakStartedDate = &started
	}
	if pet.ChargeStreak > pet.LongestStreak {
		pet.LongestStreak = pet.ChargeStreak
	}
	chargedOn := today
	pet.LastStreakDate = &chargedOn
}

func applyCharge(pet *petmodel.Pet, energy, xp int, now time.Time) {
	pet.LastChargedAt = now
	pet.EnergyPercent = energy
	pet.EnergyUpdatedAt = now
	pet.XP += xp
	pet.UpdatedAt = now
}

func chargeEvents(userID, petID uuid.UUID, dailyReward int, localDate, now time.Time) []progressionmodel.XPEvent {
	chargeID := uuid.New()
	events := []progressionmodel.XPEvent{{
		ID: chargeID, UserID: userID, PetID: petID, Source: "charge",
		SourceKey: "charge:" + chargeID.String(), Amount: rules.ChargeXPAmount,
		OccurredAt: now, LocalDate: localDate,
	}}
	if dailyReward > 0 {
		events = append(events, progressionmodel.XPEvent{
			ID: uuid.New(), UserID: userID, PetID: petID, Source: "daily_reward",
			SourceKey: dailyRewardKey(localDate), Amount: dailyReward,
			OccurredAt: now, LocalDate: localDate,
		})
	}
	return events
}

func dailyRewardKey(localDate time.Time) string {
	return fmt.Sprintf("daily_reward:%s", localDate.Format(time.DateOnly))
}
