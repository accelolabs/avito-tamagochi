package integration_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	leaderrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/repository"
	leaderservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/service"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/seed"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	taskservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/service"
	"github.com/google/uuid"
)

func TestChargeAwardsDailyRewardAndTracksStreak(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	dayOne := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	newService := func(now time.Time) petservice.Service {
		return petservice.New(
			db, petrepository.New(db), progressionrepository.New(db), rewardrepository.New(db),
			fixedClock{now: now}, nil,
		)
	}

	first, err := newService(dayOne).ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("first charge: %v", err)
	}
	if first.Pet.XP != 20 || first.BaseChargeXP != 10 || first.DailyRewardXP != 10 || first.TotalXPAwarded != 20 {
		t.Fatalf("first charge = %+v, want 20 total XP", first)
	}

	repeated, err := newService(dayOne).ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("repeated charge: %v", err)
	}
	if repeated.TotalXPAwarded != 0 || queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1`, userID) != 2 {
		t.Fatalf("repeated charge = %+v, want zero award and two events", repeated)
	}

	second, err := newService(dayOne.Add(24*time.Hour)).ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("second-day charge: %v", err)
	}
	if second.Pet.XP != 45 || second.DailyRewardXP != 15 || second.TotalXPAwarded != 25 {
		t.Fatalf("second-day charge = %+v, want 25 awarded XP", second)
	}
	streak, err := newService(dayOne.Add(24*time.Hour)).GetStreak(context.Background(), userID)
	if err != nil {
		t.Fatalf("get streak: %v", err)
	}
	if streak.CurrentStreak != 2 || streak.LongestStreak != 2 || streak.NextDailyRewardXP != 20 {
		t.Fatalf("streak = %+v, want current 2, longest 2, next reward 20", streak)
	}

	expired, err := newService(dayOne.Add(72*time.Hour)).GetStreak(context.Background(), userID)
	if err != nil {
		t.Fatalf("get expired streak: %v", err)
	}
	if expired.CurrentStreak != 0 || expired.LongestStreak != 2 || expired.NextDailyRewardXP != 10 {
		t.Fatalf("expired streak = %+v, want current 0, longest 2, next reward 10", expired)
	}
}

func TestDeadPetTaskResetsProgressAndPreservesXPHistory(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	petID := insertPet(t, db, userID, 250, now.Add(-49*time.Hour))
	insertReward(t, db, userID, now.Add(-24*time.Hour))
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO xp_events (id, user_id, pet_id, source, source_key, amount, occurred_at, local_date)
		VALUES ($1, $2, $3, 'task', 'history', 250, $4, $5)
	`, uuid.New(), userID, petID, now.Add(-24*time.Hour), now.Format(time.DateOnly))
	if err != nil {
		t.Fatalf("insert XP history: %v", err)
	}

	tasks, err := taskrepository.New(db).GetTodayProgress(context.Background(), userID, now)
	if err != nil || len(tasks) == 0 {
		t.Fatalf("get available tasks: %v, count %d", err, len(tasks))
	}
	service := taskservice.New(
		db, taskrepository.New(db), petrepository.New(db), progressionrepository.New(db),
		rewardrepository.New(db), fixedClock{now: now}, nil,
	)
	err = service.ApplyAction(context.Background(), userID, tasks[0].TaskType)
	if !errors.Is(err, gameerrors.ErrPetDead) {
		t.Fatalf("dead pet task error = %v, want pet dead", err)
	}
	if xp, _ := queryPet(t, db, userID); xp != 0 {
		t.Fatalf("pet XP = %d after death, want 0", xp)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM user_rewards WHERE user_id = $1`, userID); got != 0 {
		t.Fatalf("available rewards = %d after death, want 0", got)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1`, userID); got != 1 {
		t.Fatalf("XP history count = %d after death, want 1", got)
	}
}

func TestSeedFillsEveryLeaderboardIdempotently(t *testing.T) {
	db := testDatabase(t)
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE email LIKE 'leaderboard-seed-%@example.invalid'`)
	})

	for range 2 {
		result, err := seed.Seed(context.Background(), db, now)
		if err != nil {
			t.Fatalf("seed leaderboards: %v", err)
		}
		if result.Users != 10 {
			t.Fatalf("seeded users = %d, want 10", result.Users)
		}
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE source = 'seed'`); got != 10 {
		t.Fatalf("seed XP events = %d, want 10 after rerun", got)
	}

	service := leaderservice.NewWithClock(leaderrepository.New(db), fixedClock{now: now})
	all, err := service.GetAll(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("all-time leaderboard: %v", err)
	}
	week, err := service.GetWeekly(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("weekly leaderboard: %v", err)
	}
	month, err := service.GetMonthly(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("monthly leaderboard: %v", err)
	}
	streak, err := service.GetStreak(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("streak leaderboard: %v", err)
	}

	for name, entries := range map[string][]string{
		"all": namesFromXPEntries(all.Entries), "week": namesFromXPEntries(week.Entries),
		"month": namesFromXPEntries(month.Entries), "streak": namesFromStreakEntries(streak.Entries),
	} {
		if len(entries) != 10 {
			t.Fatalf("%s leaderboard has %d entries, want 10", name, len(entries))
		}
		for _, displayName := range entries {
			if !strings.HasPrefix(displayName, "Seed_") {
				t.Fatalf("%s leaderboard contains non-seed user %q", name, displayName)
			}
		}
	}
}

func namesFromXPEntries(entries []leadermodel.XPEntry) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.DisplayName)
	}
	return result
}

func namesFromStreakEntries(entries []leadermodel.StreakEntry) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.DisplayName)
	}
	return result
}
