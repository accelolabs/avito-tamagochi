package integration_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	rewardservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/service"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	taskservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/service"
)

const concurrentRequests = 16

func TestConcurrentChargesSerializeEnergyAndDailyReward(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 0, now)
	setPetEnergy(t, db, userID, 0, now)
	service := newPetService(db, now)

	runConcurrent(t, func() error {
		_, err := service.ChargePet(context.Background(), userID)
		return err
	})

	var xp, energy int
	if err := db.QueryRow(`SELECT xp, energy_percent FROM pets WHERE user_id = $1`, userID).Scan(&xp, &energy); err != nil {
		t.Fatalf("read pet: %v", err)
	}
	if xp != 20 || energy != 100 {
		t.Fatalf("pet after concurrent charges = (xp=%d, energy=%d), want (20, 100)", xp, energy)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1 AND source = 'daily_reward'`, userID); got != 1 {
		t.Fatalf("daily reward events = %d, want 1", got)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1 AND source = 'charge'`, userID); got != 5 {
		t.Fatalf("charge XP events = %d, want 5", got)
	}
}

func TestConcurrentPetActionsAwardOnce(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 0, now)
	setPetEnergy(t, db, userID, 100, now)
	service := newPetService(db, now)

	runConcurrent(t, func() error {
		_, err := service.Pet(context.Background(), userID)
		return err
	})

	if xp, _ := queryPet(t, db, userID); xp != 5 {
		t.Fatalf("pet XP = %d, want 5", xp)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1 AND source = 'pet'`, userID); got != 1 {
		t.Fatalf("pet XP events = %d, want 1", got)
	}
}

func TestConcurrentTaskActionsCompleteAndAwardOnce(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 0, now)
	setPetEnergy(t, db, userID, 100, now)
	service := taskservice.New(
		db, taskrepository.New(db), petrepository.New(db), progressionrepository.New(db),
		rewardrepository.New(db), fixedClock{now: now}, nil,
	)
	tasks, err := service.GetTodayTasks(context.Background(), userID)
	if err != nil || len(tasks) == 0 {
		t.Fatalf("get tasks: count=%d error=%v", len(tasks), err)
	}
	task := tasks[0]

	runConcurrent(t, func() error {
		return service.ApplyAction(context.Background(), userID, task.TaskType)
	})

	if xp, _ := queryPet(t, db, userID); xp != task.XPReward {
		t.Fatalf("pet XP = %d, want %d", xp, task.XPReward)
	}
	if got := queryInt(t, db, `SELECT progress FROM task_progress WHERE user_id = $1 AND task_type = $2`, userID, task.TaskType); got != task.RequiredCount {
		t.Fatalf("task progress = %d, want %d", got, task.RequiredCount)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1 AND source = 'task'`, userID); got != 1 {
		t.Fatalf("task XP events = %d, want 1", got)
	}
}

func TestConcurrentRewardUseSucceedsOnce(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	rewardID := insertReward(t, db, userID, time.Now().UTC())
	service := rewardservice.New(db, rewardrepository.New(db), nil)
	var successes atomic.Int32

	runConcurrent(t, func() error {
		_, err := service.UseReward(context.Background(), userID, rewardID)
		if err == nil {
			successes.Add(1)
			return nil
		}
		if errors.Is(err, gameerrors.ErrRewardAlreadyUsed) {
			return nil
		}
		return err
	})

	if successes.Load() != 1 {
		t.Fatalf("successful reward uses = %d, want 1", successes.Load())
	}
	status, usedAt := queryReward(t, db, rewardID)
	if status != "used" || usedAt == nil {
		t.Fatalf("reward = (status=%q, usedAt=%v), want used", status, usedAt)
	}
}

func runConcurrent(t *testing.T, operation func() error) {
	t.Helper()
	start := make(chan struct{})
	errorsChannel := make(chan error, concurrentRequests)
	var workers sync.WaitGroup
	for range concurrentRequests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			errorsChannel <- operation()
		}()
	}
	close(start)
	workers.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent operation: %v", err)
		}
	}
}
