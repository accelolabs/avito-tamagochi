package service

import (
	"context"
	"database/sql"
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
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetTodayTasks(context.Context, uuid.UUID) ([]taskmodel.Progress, error)
	ApplyAction(context.Context, uuid.UUID, taskmodel.Type) error
}

type service struct {
	db         *sql.DB
	repo       taskrepository.Repository
	petRepo    petrepository.Repository
	xpRepo     progressionrepository.Repository
	rewardRepo rewardrepository.Repository
	clock      clock.Clock
	notify     notifier.Notifier
}

func New(db *sql.DB, repo taskrepository.Repository, petRepo petrepository.Repository, xpRepo progressionrepository.Repository, rewardRepo rewardrepository.Repository, now clock.Clock, notify notifier.Notifier) Service {
	if now == nil {
		now = clock.RealClock{}
	}
	return &service{db: db, repo: repo, petRepo: petRepo, xpRepo: xpRepo, rewardRepo: rewardRepo, clock: now, notify: notify}
}

func (s *service) GetTodayTasks(ctx context.Context, userID uuid.UUID) ([]taskmodel.Progress, error) {
	return s.repo.GetTodayProgress(ctx, userID, clock.MoscowDate(s.clock.Now()))
}

func (s *service) ApplyAction(ctx context.Context, userID uuid.UUID, taskType taskmodel.Type) error {
	if rules.TaskXP(taskType) == 0 {
		return gameerrors.ErrInvalidAction
	}
	now := s.clock.Now().UTC()
	localDate := clock.MoscowDate(now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	progress, err := s.repo.GetProgressForUpdate(ctx, tx, userID, localDate, taskType)
	if err != nil {
		return err
	}
	pet, err := s.petRepo.GetOrCreateForUpdate(ctx, tx, userID, petmodel.Pet{ID: uuid.New(), UserID: userID, LastChargedAt: now.Add(-24 * time.Hour), CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return err
	}
	if rules.IsDead(pet.LastChargedAt, now) {
		if err := s.petRepo.ResetAfterDeath(ctx, tx, pet.ID, userID); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if s.notify != nil {
			s.notify.NotifyUser(userID, "pet_updated")
			s.notify.NotifyUser(userID, "rewards_updated")
		}
		return gameerrors.ErrPetDead
	}
	if progress.Completed {
		return tx.Commit()
	}
	oldLevel := rules.LevelFromXP(pet.XP)
	levelReached := false
	progress.Progress++
	if progress.Progress >= progress.RequiredCount {
		progress.Progress = progress.RequiredCount
		completedAt := now
		progress.CompletedAt = &completedAt
		progress.Completed = true
		pet.XP += rules.TaskXP(taskType)
		pet.UpdatedAt = now
		if err := s.petRepo.Update(ctx, tx, *pet); err != nil {
			return err
		}
		event := progressionmodel.XPEvent{ID: uuid.New(), UserID: userID, PetID: pet.ID, Source: "task", SourceKey: fmt.Sprintf("task:%s:%s", taskType, localDate.Format("2006-01-02")), Amount: rules.TaskXP(taskType), OccurredAt: now, LocalDate: localDate}
		if err := s.xpRepo.CreateXPEvent(ctx, tx, event); err != nil {
			return err
		}
		newLevel := rules.LevelFromXP(pet.XP)
		levelReached = newLevel > oldLevel
		for level := oldLevel + 1; level <= newLevel; level++ {
			if err := s.rewardRepo.UnlockForLevel(ctx, tx, userID, level, rules.RewardTypeForLevel(level), now); err != nil {
				return err
			}
		}
	}
	if err := s.repo.SaveProgress(ctx, tx, *progress); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if s.notify != nil {
		s.notify.NotifyUser(userID, "tasks_updated")
		if progress.Completed {
			s.notify.NotifyUser(userID, "pet_updated")
			if levelReached {
				s.notify.NotifyUser(userID, "rewards_updated")
			}
		}
	}
	return nil
}
