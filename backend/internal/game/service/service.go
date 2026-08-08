package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/pet"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression"
	"github.com/google/uuid"
)

type Service interface {
	GetPet(context.Context, uuid.UUID) (*pet.Stats, error)
	ChargePet(context.Context, uuid.UUID) (*pet.Stats, error)
}

type service struct {
	db       *sql.DB
	petRepo  pet.Repository
	xpRepo   progression.Repository
	clock    game.Clock
	notifier game.Notifier
}

func NewService(db *sql.DB, petRepo pet.Repository, xpRepo progression.Repository, clock game.Clock, notifier game.Notifier) Service {
	if clock == nil {
		clock = game.RealClock{}
	}
	return &service{db: db, petRepo: petRepo, xpRepo: xpRepo, clock: clock, notifier: notifier}
}

func (s *service) GetPet(ctx context.Context, userID uuid.UUID) (*pet.Stats, error) {
	now := s.clock.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	value, err := s.getOrCreatePet(ctx, tx, userID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return stats(value, now), nil
}

func (s *service) ChargePet(ctx context.Context, userID uuid.UUID) (*pet.Stats, error) {
	now := s.clock.Now().UTC()
	localDate := progression.MoscowDate(now)
	sourceKey := fmt.Sprintf("charge:%s", localDate.Format("2006-01-02"))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	value, err := s.getOrCreatePet(ctx, tx, userID, now)
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
		return stats(value, now), nil
	}

	value.LastChargedAt = now
	value.XP += progression.ChargeXPAmount
	value.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *value); err != nil {
		return nil, err
	}

	event := progression.XPEvent{
		ID:         uuid.New(),
		UserID:     userID,
		PetID:      value.ID,
		Source:     "charge",
		SourceKey:  sourceKey,
		Amount:     progression.ChargeXPAmount,
		OccurredAt: now,
		LocalDate:  localDate,
	}
	if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if s.notifier != nil {
		s.notifier.NotifyUser(userID, "pet_updated")
	}
	return stats(value, now), nil
}

func (s *service) getOrCreatePet(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time) (*pet.Pet, error) {
	initial := pet.Pet{
		ID:            uuid.New(),
		UserID:        userID,
		LastChargedAt: now.Add(-24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, initial)
}

func stats(value *pet.Pet, now time.Time) *pet.Stats {
	return &pet.Stats{
		XP:            value.XP,
		Level:         progression.LevelFromXP(value.XP),
		Energy:        progression.EnergyPercent(value.LastChargedAt, now),
		LastChargedAt: value.LastChargedAt,
	}
}
