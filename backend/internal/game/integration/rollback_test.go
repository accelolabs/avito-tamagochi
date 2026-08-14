package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	rewardservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/service"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	taskservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/service"
	"github.com/google/uuid"
)

var errInjectedWrite = errors.New("injected failure after write")

var rollbackTestTime = time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)

type chargeRepositories struct {
	pet    petrepository.Repository
	xp     progressionrepository.Repository
	reward rewardrepository.Repository
}

func TestChargePetRollsBackIntermediateWrites(t *testing.T) {
	tests := []struct {
		name    string
		wrap    func(chargeRepositories) chargeRepositories
		wantErr error
	}{
		{
			name: "after pet update",
			wrap: func(repositories chargeRepositories) chargeRepositories {
				repositories.pet = failAfterPetUpdateRepository{Repository: repositories.pet, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "after XP event insert",
			wrap: func(repositories chargeRepositories) chargeRepositories {
				repositories.xp = failAfterXPEventRepository{Repository: repositories.xp, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "after reward unlock",
			wrap: func(repositories chargeRepositories) chargeRepositories {
				repositories.reward = failAfterRewardUnlockRepository{Repository: repositories.reward, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "at commit",
			wrap: func(repositories chargeRepositories) chargeRepositories {
				repositories.reward = rollbackAfterRewardUnlockRepository{Repository: repositories.reward}
				return repositories
			},
			wantErr: sql.ErrTxDone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := testDatabase(t)
			userID := insertTestUser(t, db)
			lastChargedAt := rollbackTestTime.Add(-24 * time.Hour)
			insertPet(t, db, userID, 90, lastChargedAt)

			repositories := test.wrap(chargeRepositories{
				pet:    petrepository.New(db),
				xp:     progressionrepository.New(db),
				reward: rewardrepository.New(db),
			})
			notifier := &recordingNotifier{}
			service := petservice.New(db, repositories.pet, repositories.xp, repositories.reward, fixedClock{now: rollbackTestTime}, notifier)

			_, err := service.ChargePet(context.Background(), userID)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ChargePet() error = %v, want %v", err, test.wantErr)
			}

			xp, chargedAt := queryPet(t, db, userID)
			if xp != 90 || !chargedAt.Equal(lastChargedAt) {
				t.Fatalf("pet state = (xp=%d, lastChargedAt=%s), want (xp=90, lastChargedAt=%s)", xp, chargedAt, lastChargedAt)
			}
			if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1`, userID); got != 0 {
				t.Fatalf("XP event count = %d, want 0", got)
			}
			if got := queryInt(t, db, `SELECT COUNT(*) FROM user_rewards WHERE user_id = $1`, userID); got != 0 {
				t.Fatalf("reward count = %d, want 0", got)
			}
			notifier.assert(t, userID)
		})
	}
}

func TestChargePetCommitsBeforeNotifying(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	insertPet(t, db, userID, 90, rollbackTestTime.Add(-24*time.Hour))
	notifier := &recordingNotifier{observe: committedChargeState(db, userID)}
	service := petservice.New(
		db,
		petrepository.New(db),
		progressionrepository.New(db),
		rewardrepository.New(db),
		fixedClock{now: rollbackTestTime},
		notifier,
	)

	result, err := service.ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("ChargePet() error = %v", err)
	}
	if result.Pet.XP != 102 || result.Pet.Level != 2 || result.Pet.Energy != 70 {
		t.Fatalf("ChargePet() result = %+v, want XP 102, level 2, energy 70", result)
	}
	if result.BaseChargeXP != 2 || result.DailyRewardXP != 10 || result.TotalXPAwarded != 12 {
		t.Fatalf("ChargePet() award = %+v, want 2 base, 10 daily, 12 total", result)
	}
	notifier.assert(t, userID, "pet_updated", "rewards_updated")
}

func TestChargePetRollsBackImplicitPetCreation(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	notifier := &recordingNotifier{}
	xpRepository := failAfterXPEventRepository{Repository: progressionrepository.New(db), err: errInjectedWrite}
	service := petservice.New(
		db,
		petrepository.New(db),
		xpRepository,
		rewardrepository.New(db),
		fixedClock{now: rollbackTestTime},
		notifier,
	)

	_, err := service.ChargePet(context.Background(), userID)
	if !errors.Is(err, errInjectedWrite) {
		t.Fatalf("ChargePet() error = %v, want %v", err, errInjectedWrite)
	}
	assertNoGameState(t, db, userID)
	notifier.assert(t, userID)
}

type taskRepositories struct {
	task   taskrepository.Repository
	pet    petrepository.Repository
	xp     progressionrepository.Repository
	reward rewardrepository.Repository
}

func TestTaskCompletionRollsBackIntermediateWrites(t *testing.T) {
	tests := []struct {
		name    string
		wrap    func(taskRepositories) taskRepositories
		wantErr error
	}{
		{
			name: "after pet update",
			wrap: func(repositories taskRepositories) taskRepositories {
				repositories.pet = failAfterPetUpdateRepository{Repository: repositories.pet, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "after XP event insert",
			wrap: func(repositories taskRepositories) taskRepositories {
				repositories.xp = failAfterXPEventRepository{Repository: repositories.xp, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "after reward unlock",
			wrap: func(repositories taskRepositories) taskRepositories {
				repositories.reward = failAfterRewardUnlockRepository{Repository: repositories.reward, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "after task progress save",
			wrap: func(repositories taskRepositories) taskRepositories {
				repositories.task = failAfterTaskSaveRepository{Repository: repositories.task, err: errInjectedWrite}
				return repositories
			},
			wantErr: errInjectedWrite,
		},
		{
			name: "at commit",
			wrap: func(repositories taskRepositories) taskRepositories {
				repositories.task = rollbackAfterTaskSaveRepository{Repository: repositories.task}
				return repositories
			},
			wantErr: sql.ErrTxDone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := testDatabase(t)
			userID := insertTestUser(t, db)
			insertPet(t, db, userID, 60, rollbackTestTime.Add(-24*time.Hour))
			repositories := test.wrap(taskRepositories{
				task:   taskrepository.New(db),
				pet:    petrepository.New(db),
				xp:     progressionrepository.New(db),
				reward: rewardrepository.New(db),
			})
			notifier := &recordingNotifier{}
			service := taskservice.New(db, repositories.task, repositories.pet, repositories.xp, repositories.reward, fixedClock{now: rollbackTestTime}, notifier)

			err := service.ApplyAction(context.Background(), userID, taskmodel.CreateListing)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ApplyAction() error = %v, want %v", err, test.wantErr)
			}

			xp, _ := queryPet(t, db, userID)
			if xp != 60 {
				t.Fatalf("pet XP = %d, want 60", xp)
			}
			if got := queryInt(t, db, `SELECT COUNT(*) FROM task_progress WHERE user_id = $1`, userID); got != 0 {
				t.Fatalf("task progress count = %d, want 0", got)
			}
			if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1`, userID); got != 0 {
				t.Fatalf("XP event count = %d, want 0", got)
			}
			if got := queryInt(t, db, `SELECT COUNT(*) FROM user_rewards WHERE user_id = $1`, userID); got != 0 {
				t.Fatalf("reward count = %d, want 0", got)
			}
			notifier.assert(t, userID)
		})
	}
}

func TestTaskCompletionCommitsBeforeNotifying(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	insertPet(t, db, userID, 60, rollbackTestTime.Add(-24*time.Hour))
	localDate := clock.MoscowDate(rollbackTestTime)
	notifier := &recordingNotifier{observe: committedTaskState(db, userID, localDate)}
	service := taskservice.New(
		db,
		taskrepository.New(db),
		petrepository.New(db),
		progressionrepository.New(db),
		rewardrepository.New(db),
		fixedClock{now: rollbackTestTime},
		notifier,
	)

	if err := service.ApplyAction(context.Background(), userID, taskmodel.CreateListing); err != nil {
		t.Fatalf("ApplyAction() error = %v", err)
	}
	notifier.assert(t, userID, "tasks_updated", "pet_updated", "rewards_updated")
}

func TestTaskActionRollsBackImplicitPetAndProgressCreation(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	notifier := &recordingNotifier{}
	taskRepository := failAfterTaskSaveRepository{Repository: taskrepository.New(db), err: errInjectedWrite}
	service := taskservice.New(
		db,
		taskRepository,
		petrepository.New(db),
		progressionrepository.New(db),
		rewardrepository.New(db),
		fixedClock{now: rollbackTestTime},
		notifier,
	)

	err := service.ApplyAction(context.Background(), userID, taskmodel.CreateListing)
	if !errors.Is(err, errInjectedWrite) {
		t.Fatalf("ApplyAction() error = %v, want %v", err, errInjectedWrite)
	}
	assertNoGameState(t, db, userID)
	notifier.assert(t, userID)
}

func TestUseRewardRollsBackIntermediateWrite(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	rewardID := insertReward(t, db, userID, rollbackTestTime)
	repository := failAfterRewardUseRepository{Repository: rewardrepository.New(db), err: errInjectedWrite}
	notifier := &recordingNotifier{}
	service := rewardservice.New(db, repository, notifier)

	_, err := service.UseReward(context.Background(), userID, rewardID)
	if !errors.Is(err, errInjectedWrite) {
		t.Fatalf("UseReward() error = %v, want %v", err, errInjectedWrite)
	}
	assertRewardAvailable(t, db, rewardID)
	notifier.assert(t, userID)
}

func TestUseRewardCommitFailureRollsBackAndDoesNotNotify(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	rewardID := insertReward(t, db, userID, rollbackTestTime)
	repository := rollbackAfterRewardUseRepository{Repository: rewardrepository.New(db)}
	notifier := &recordingNotifier{}
	service := rewardservice.New(db, repository, notifier)

	_, err := service.UseReward(context.Background(), userID, rewardID)
	if !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("UseReward() error = %v, want %v", err, sql.ErrTxDone)
	}
	assertRewardAvailable(t, db, rewardID)
	notifier.assert(t, userID)
}

func TestUseRewardCommitsBeforeNotifying(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	rewardID := insertReward(t, db, userID, rollbackTestTime)
	notifier := &recordingNotifier{observe: committedRewardState(db, rewardID)}
	service := rewardservice.New(db, rewardrepository.New(db), notifier)

	value, err := service.UseReward(context.Background(), userID, rewardID)
	if err != nil {
		t.Fatalf("UseReward() error = %v", err)
	}
	if value.Status != "used" || value.UsedAt == nil {
		t.Fatalf("UseReward() = %+v, want used reward", value)
	}
	notifier.assert(t, userID, "rewards_updated")
}

func committedChargeState(db *sql.DB, userID uuid.UUID) func(uuid.UUID, string) error {
	return func(_ uuid.UUID, _ string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var xp, events, rewards int
		err := db.QueryRowContext(ctx, `
			SELECT p.xp,
			       (SELECT COUNT(*) FROM xp_events WHERE user_id = $1),
			       (SELECT COUNT(*) FROM user_rewards WHERE user_id = $1)
			FROM pets p WHERE p.user_id = $1
		`, userID).Scan(&xp, &events, &rewards)
		if err != nil {
			return err
		}
		if xp != 102 || events != 2 || rewards != 1 {
			return fmt.Errorf("charge state = (xp=%d, events=%d, rewards=%d), want (102, 2, 1)", xp, events, rewards)
		}
		return nil
	}
}

func committedTaskState(db *sql.DB, userID uuid.UUID, localDate time.Time) func(uuid.UUID, string) error {
	return func(_ uuid.UUID, _ string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var xp, events, rewards, progress int
		var completedAt *time.Time
		err := db.QueryRowContext(ctx, `
			SELECT p.xp,
			       (SELECT COUNT(*) FROM xp_events WHERE user_id = $1),
			       (SELECT COUNT(*) FROM user_rewards WHERE user_id = $1),
			       tp.progress,
			       tp.completed_at
			FROM pets p
			JOIN task_progress tp ON tp.user_id = p.user_id
			WHERE p.user_id = $1 AND tp.local_date = $2 AND tp.task_type = $3
		`, userID, localDate.Format("2006-01-02"), taskmodel.CreateListing).Scan(&xp, &events, &rewards, &progress, &completedAt)
		if err != nil {
			return err
		}
		if xp != 100 || events != 1 || rewards != 1 || progress != 1 || completedAt == nil {
			return fmt.Errorf("task state = (xp=%d, events=%d, rewards=%d, progress=%d, completed=%t), want (100, 1, 1, 1, true)", xp, events, rewards, progress, completedAt != nil)
		}
		return nil
	}
}

func committedRewardState(db *sql.DB, rewardID uuid.UUID) func(uuid.UUID, string) error {
	return func(_ uuid.UUID, _ string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var status string
		var usedAt *time.Time
		if err := db.QueryRowContext(ctx, `SELECT status, used_at FROM user_rewards WHERE id = $1`, rewardID).Scan(&status, &usedAt); err != nil {
			return err
		}
		if status != "used" || usedAt == nil {
			return fmt.Errorf("reward state = (status=%s, usedAt=%v), want used with timestamp", status, usedAt)
		}
		return nil
	}
}

func assertRewardAvailable(t *testing.T, db *sql.DB, rewardID uuid.UUID) {
	t.Helper()
	status, usedAt := queryReward(t, db, rewardID)
	if status != "available" || usedAt != nil {
		t.Fatalf("reward state = (status=%s, usedAt=%v), want available with nil timestamp", status, usedAt)
	}
}

func assertNoGameState(t *testing.T, db *sql.DB, userID uuid.UUID) {
	t.Helper()
	for table, query := range map[string]string{
		"pets":          `SELECT COUNT(*) FROM pets WHERE user_id = $1`,
		"task_progress": `SELECT COUNT(*) FROM task_progress WHERE user_id = $1`,
		"xp_events":     `SELECT COUNT(*) FROM xp_events WHERE user_id = $1`,
		"user_rewards":  `SELECT COUNT(*) FROM user_rewards WHERE user_id = $1`,
	} {
		if got := queryInt(t, db, query, userID); got != 0 {
			t.Fatalf("%s row count = %d, want 0", table, got)
		}
	}
}
