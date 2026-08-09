package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	petrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	progressionmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/model"
	progressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	rewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/migrations"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const gameTestDatabaseEnvironment = "GAME_TEST_DATABASE_URL"

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

var _ clock.Clock = fixedClock{}

type notification struct {
	userID uuid.UUID
	event  string
}

type recordingNotifier struct {
	mu      sync.Mutex
	items   []notification
	observe func(uuid.UUID, string) error
	errors  []error
}

func (n *recordingNotifier) NotifyUser(userID uuid.UUID, event string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.observe != nil {
		if err := n.observe(userID, event); err != nil {
			n.errors = append(n.errors, err)
		}
	}
	n.items = append(n.items, notification{userID: userID, event: event})
}

func (n *recordingNotifier) assert(t *testing.T, userID uuid.UUID, events ...string) {
	t.Helper()
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.errors) > 0 {
		t.Fatalf("notification observed uncommitted state: %v", n.errors)
	}
	got := make([]string, 0, len(n.items))
	for _, item := range n.items {
		if item.userID != userID {
			t.Fatalf("notification user ID = %s, want %s", item.userID, userID)
		}
		got = append(got, item.event)
	}
	if len(got) != len(events) {
		t.Fatalf("notifications = %v, want %v", got, events)
	}
	for index := range got {
		if got[index] != events[index] {
			t.Fatalf("notifications = %v, want %v", got, events)
		}
	}
}

func testDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv(gameTestDatabaseEnvironment)
	if databaseURL == "" {
		t.Skipf("%s is not set", gameTestDatabaseEnvironment)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	migrationsDirectory := filepath.Join("..", "..", "..", "migrations")
	if err := migrations.Apply(ctx, db, migrationsDirectory); err != nil {
		t.Fatalf("apply test migrations: %v", err)
	}
	return db
}

func insertTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, fmt.Sprintf("game-rollback-%s@example.com", userID), "Rollback_Test", "test-password-hash", time.Now().UTC())
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})
	return userID
}

func insertPet(t *testing.T, db *sql.DB, userID uuid.UUID, xp int, lastChargedAt time.Time) uuid.UUID {
	t.Helper()
	petID := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO pets (id, user_id, xp, last_charged_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, petID, userID, xp, lastChargedAt, lastChargedAt)
	if err != nil {
		t.Fatalf("insert pet: %v", err)
	}
	return petID
}

func insertReward(t *testing.T, db *sql.DB, userID uuid.UUID, now time.Time) uuid.UUID {
	t.Helper()
	rewardID := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO user_rewards (id, user_id, level, type, status, unlocked_at)
		VALUES ($1, $2, 2, 'promotion', 'available', $3)
	`, rewardID, userID, now)
	if err != nil {
		t.Fatalf("insert reward: %v", err)
	}
	return rewardID
}

func queryInt(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var value int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query integer: %v", err)
	}
	return value
}

func queryPet(t *testing.T, db *sql.DB, userID uuid.UUID) (int, time.Time) {
	t.Helper()
	var xp int
	var lastChargedAt time.Time
	if err := db.QueryRowContext(context.Background(), `SELECT xp, last_charged_at FROM pets WHERE user_id = $1`, userID).Scan(&xp, &lastChargedAt); err != nil {
		t.Fatalf("query pet: %v", err)
	}
	return xp, lastChargedAt
}

func queryReward(t *testing.T, db *sql.DB, rewardID uuid.UUID) (string, *time.Time) {
	t.Helper()
	var status string
	var usedAt *time.Time
	if err := db.QueryRowContext(context.Background(), `SELECT status, used_at FROM user_rewards WHERE id = $1`, rewardID).Scan(&status, &usedAt); err != nil {
		t.Fatalf("query reward: %v", err)
	}
	return status, usedAt
}

type failAfterPetUpdateRepository struct {
	petrepository.Repository
	err error
}

func (r failAfterPetUpdateRepository) Update(ctx context.Context, tx *sql.Tx, value petmodel.Pet) error {
	if err := r.Repository.Update(ctx, tx, value); err != nil {
		return err
	}
	return r.err
}

type failAfterXPEventRepository struct {
	progressionrepository.Repository
	err error
}

func (r failAfterXPEventRepository) CreateXPEvent(ctx context.Context, tx *sql.Tx, event progressionmodel.XPEvent) error {
	if err := r.Repository.CreateXPEvent(ctx, tx, event); err != nil {
		return err
	}
	return r.err
}

type failAfterRewardUnlockRepository struct {
	rewardrepository.Repository
	err error
}

func (r failAfterRewardUnlockRepository) UnlockForLevel(ctx context.Context, tx *sql.Tx, userID uuid.UUID, level int, rewardType rewardmodel.Type, unlockedAt time.Time) error {
	if err := r.Repository.UnlockForLevel(ctx, tx, userID, level, rewardType, unlockedAt); err != nil {
		return err
	}
	return r.err
}

type rollbackAfterRewardUnlockRepository struct {
	rewardrepository.Repository
}

func (r rollbackAfterRewardUnlockRepository) UnlockForLevel(ctx context.Context, tx *sql.Tx, userID uuid.UUID, level int, rewardType rewardmodel.Type, unlockedAt time.Time) error {
	if err := r.Repository.UnlockForLevel(ctx, tx, userID, level, rewardType, unlockedAt); err != nil {
		return err
	}
	return tx.Rollback()
}

type failAfterTaskSaveRepository struct {
	taskrepository.Repository
	err error
}

func (r failAfterTaskSaveRepository) SaveProgress(ctx context.Context, tx *sql.Tx, progress taskmodel.Progress) error {
	if err := r.Repository.SaveProgress(ctx, tx, progress); err != nil {
		return err
	}
	return r.err
}

type rollbackAfterTaskSaveRepository struct {
	taskrepository.Repository
}

func (r rollbackAfterTaskSaveRepository) SaveProgress(ctx context.Context, tx *sql.Tx, progress taskmodel.Progress) error {
	if err := r.Repository.SaveProgress(ctx, tx, progress); err != nil {
		return err
	}
	return tx.Rollback()
}

type failAfterRewardUseRepository struct {
	rewardrepository.Repository
	err error
}

func (r failAfterRewardUseRepository) Use(ctx context.Context, tx *sql.Tx, userID, rewardID uuid.UUID, usedAt time.Time) (*rewardmodel.UserReward, error) {
	value, err := r.Repository.Use(ctx, tx, userID, rewardID, usedAt)
	if err != nil {
		return nil, err
	}
	return value, r.err
}

type rollbackAfterRewardUseRepository struct {
	rewardrepository.Repository
}

func (r rollbackAfterRewardUseRepository) Use(ctx context.Context, tx *sql.Tx, userID, rewardID uuid.UUID, usedAt time.Time) (*rewardmodel.UserReward, error) {
	value, err := r.Repository.Use(ctx, tx, userID, rewardID, usedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Rollback(); err != nil {
		return nil, err
	}
	return value, nil
}
