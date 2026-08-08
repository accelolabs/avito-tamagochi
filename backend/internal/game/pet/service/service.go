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
	progression "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/service"
	"github.com/google/uuid"
)

type Service interface {
	GetPet(context.Context, uuid.UUID) (*petmodel.Stats, error)
	ChargePet(context.Context, uuid.UUID) (*petmodel.Stats, error)
}

type service struct {
	db       *sql.DB
	petRepo  petrepository.Repository
	xpRepo   progressionrepository.Repository
	clock    clock.Clock
	progress progression.Service
	notify   notifier.Notifier
}

func New(db *sql.DB, petRepo petrepository.Repository, xpRepo progressionrepository.Repository, now clock.Clock, progress progression.Service, notify notifier.Notifier) Service {
	if now == nil {
		now = clock.RealClock{}
	}
	if progress == nil {
		progress = progression.New()
	}
	return &service{db: db, petRepo: petRepo, xpRepo: xpRepo, clock: now, progress: progress, notify: notify}
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
	localDate := s.progress.MoscowDate(now)
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
	value.LastChargedAt = now
	value.XP += s.progress.ChargeXP()
	value.UpdatedAt = now
	if err := s.petRepo.Update(ctx, tx, *value); err != nil {
		return nil, err
	}
	event := progressionmodel.XPEvent{
		ID: uuid.New(), UserID: userID, PetID: value.ID, Source: "charge", SourceKey: sourceKey,
		Amount: s.progress.ChargeXP(), OccurredAt: now, LocalDate: localDate,
	}
	if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if s.notify != nil {
		s.notify.NotifyUser(userID, "pet_updated")
	}
	return s.stats(value, now), nil
}

func (s *service) getOrCreate(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time) (*petmodel.Pet, error) {
	initial := petmodel.Pet{ID: uuid.New(), UserID: userID, LastChargedAt: now.Add(-24 * time.Hour), CreatedAt: now, UpdatedAt: now}
	return s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, initial)
}

func (s *service) stats(value *petmodel.Pet, now time.Time) *petmodel.Stats {
	return &petmodel.Stats{XP: value.XP, Level: s.progress.LevelFromXP(value.XP), Energy: s.progress.EnergyPercent(value.LastChargedAt, now), LastChargedAt: value.LastChargedAt}
}
