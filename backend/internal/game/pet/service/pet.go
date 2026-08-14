package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	progressionmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/model"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
	"github.com/google/uuid"
)

func (s *service) Pet(ctx context.Context, userID uuid.UUID) (*petmodel.PetActionResult, error) {
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
	if rules.IsDead(s.energy(pet, now)) {
		return nil, s.rejectDeadPet(ctx, tx, pet)
	}

	sourceKey := fmt.Sprintf("pet:%s", localDate.Format(time.DateOnly))
	exists, err := s.xpRepo.HasSourceKey(ctx, tx, userID, sourceKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return s.unchangedPetAction(tx, pet, now)
	}

	oldXP := pet.XP
	pet.XP += rules.PetXPAmount
	pet.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *pet); err != nil {
		return nil, err
	}
	if err := s.createXPEvents(ctx, tx, petXPEvent(userID, pet.ID, sourceKey, localDate, now)); err != nil {
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
	return &petmodel.PetActionResult{Pet: s.stats(pet, now), XPAwarded: rules.PetXPAmount}, nil
}

func (s *service) rejectDeadPet(ctx context.Context, tx *sql.Tx, pet *petmodel.Pet) error {
	reset := needsDeathReset(pet)
	if reset {
		if err := s.resetAfterDeath(ctx, tx, pet); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if reset {
		s.notifyDeath(pet.UserID)
	}
	return gameerrors.ErrPetDead
}

func (s *service) unchangedPetAction(tx *sql.Tx, pet *petmodel.Pet, now time.Time) (*petmodel.PetActionResult, error) {
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &petmodel.PetActionResult{Pet: s.stats(pet, now)}, nil
}

func petXPEvent(userID, petID uuid.UUID, sourceKey string, localDate, now time.Time) progressionmodel.XPEvent {
	return progressionmodel.XPEvent{
		ID: uuid.New(), UserID: userID, PetID: petID, Source: "pet",
		SourceKey: sourceKey, Amount: rules.PetXPAmount, OccurredAt: now, LocalDate: localDate,
	}
}
