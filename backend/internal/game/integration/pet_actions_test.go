package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	"github.com/google/uuid"
)

func TestChargeAddsTwentyCapsAtFullAndResetsNotificationThresholds(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 0, now)
	setPetEnergy(t, db, userID, 90, now)
	for _, threshold := range []int{0, 5, 25, 50} {
		if _, err := db.ExecContext(context.Background(), `
			INSERT INTO energy_notification_deliveries (user_id, threshold, sent_at)
			VALUES ($1, $2, $3)
		`, userID, threshold, now); err != nil {
			t.Fatalf("insert notification threshold %d: %v", threshold, err)
		}
	}
	service := newPetService(db, now)

	first, err := service.ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("charge 90 percent pet: %v", err)
	}
	if first.Pet.Energy != 100 || first.BaseChargeXP != 2 || first.DailyRewardXP != 10 || first.TotalXPAwarded != 12 {
		t.Fatalf("first charge = %+v", first)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM energy_notification_deliveries WHERE user_id = $1`, userID); got != 0 {
		t.Fatalf("delivery thresholds after full charge = %d, want 0", got)
	}

	full, err := service.ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("charge full pet: %v", err)
	}
	if full.TotalXPAwarded != 0 || full.Pet.XP != first.Pet.XP {
		t.Fatalf("full charge = %+v, want zero award", full)
	}

	for _, threshold := range []int{0, 5, 25, 50} {
		if _, err := db.ExecContext(context.Background(), `
			INSERT INTO energy_notification_deliveries (user_id, threshold, sent_at)
			VALUES ($1, $2, $3)
		`, userID, threshold, now); err != nil {
			t.Fatalf("reinsert notification threshold %d: %v", threshold, err)
		}
	}
	setPetEnergy(t, db, userID, 10, now)
	third, err := service.ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("charge again: %v", err)
	}
	if third.Pet.Energy != 30 || third.TotalXPAwarded != 2 || third.DailyRewardXP != 0 {
		t.Fatalf("same-day charge = %+v", third)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM energy_notification_deliveries WHERE user_id = $1 AND threshold = 50`, userID); got != 1 {
		t.Fatalf("threshold 50 count = %d, want retained", got)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM energy_notification_deliveries WHERE user_id = $1`, userID); got != 1 {
		t.Fatalf("delivery thresholds after partial charge = %d, want 1", got)
	}
}

func TestChargeRevivesDeadPetAtTwentyPercent(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 90, now.Add(-48*time.Hour))
	setPetEnergy(t, db, userID, 0, now)

	result, err := newPetService(db, now).ChargePet(context.Background(), userID)
	if err != nil {
		t.Fatalf("revive pet: %v", err)
	}
	if result.Pet.Energy != 20 || result.Pet.IsDead || result.Pet.XP != 12 {
		t.Fatalf("revived pet = %+v, want 20 energy and reset XP plus 12", result.Pet)
	}
}

func TestPetActionAwardsOncePerMoscowDay(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	dayOne := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 98, dayOne)
	setPetEnergy(t, db, userID, 100, dayOne)

	first, err := newPetService(db, dayOne).Pet(context.Background(), userID)
	if err != nil {
		t.Fatalf("first pet action: %v", err)
	}
	if first.XPAwarded != 5 || first.Pet.XP != 103 || first.Pet.Energy != 100 {
		t.Fatalf("first pet action = %+v", first)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM user_rewards WHERE user_id = $1`, userID); got != 1 {
		t.Fatalf("unlocked rewards = %d, want 1", got)
	}

	repeated, err := newPetService(db, dayOne).Pet(context.Background(), userID)
	if err != nil || repeated.XPAwarded != 0 || repeated.Pet.XP != 103 {
		t.Fatalf("repeated pet action = %+v error=%v", repeated, err)
	}

	dayTwo := dayOne.Add(24 * time.Hour)
	next, err := newPetService(db, dayTwo).Pet(context.Background(), userID)
	if err != nil || next.XPAwarded != 5 || next.Pet.XP != 108 {
		t.Fatalf("next-day pet action = %+v error=%v", next, err)
	}
}

func TestPetActionRejectsDeadPetAndAppliesReset(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	insertPet(t, db, userID, 90, now.Add(-48*time.Hour))
	setPetEnergy(t, db, userID, 0, now)

	result, err := newPetService(db, now).Pet(context.Background(), userID)
	if !errors.Is(err, gameerrors.ErrPetDead) || result != nil {
		t.Fatalf("dead pet action result=%+v error=%v", result, err)
	}
	if xp, _ := queryPet(t, db, userID); xp != 0 {
		t.Fatalf("dead pet XP = %d, want reset", xp)
	}
	if got := queryInt(t, db, `SELECT COUNT(*) FROM xp_events WHERE user_id = $1 AND source = 'pet'`, userID); got != 0 {
		t.Fatalf("dead pet XP events = %d, want 0", got)
	}
}

func newPetService(db *sql.DB, now time.Time) petservice.Service {
	return petservice.New(
		db, petrepository.New(db), progressionrepository.New(db), rewardrepository.New(db),
		fixedClock{now: now}, nil,
	)
}

func setPetEnergy(t *testing.T, db *sql.DB, userID uuid.UUID, energy int, updatedAt time.Time) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		UPDATE pets SET energy_percent = $2, energy_updated_at = $3 WHERE user_id = $1
	`, userID, energy, updatedAt); err != nil {
		t.Fatalf("set pet energy: %v", err)
	}
}
