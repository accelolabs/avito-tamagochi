package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/migrations"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func TestPostgreSQLRepositorySerializesRunsAndRecordsDelivery(t *testing.T) {
	db := notificationTestDatabase(t)
	participant := insertNotificationParticipant(t, db)
	repo := New(db)

	release, acquired, err := repo.TryRunLock(context.Background())
	if err != nil || !acquired {
		t.Fatalf("first lock acquired=%v error=%v", acquired, err)
	}
	_, secondAcquired, err := repo.TryRunLock(context.Background())
	if err != nil || secondAcquired {
		t.Fatalf("second lock acquired=%v error=%v", secondAcquired, err)
	}
	release()
	releaseAgain, acquiredAgain, err := repo.TryRunLock(context.Background())
	if err != nil || !acquiredAgain {
		t.Fatalf("lock after release acquired=%v error=%v", acquiredAgain, err)
	}
	releaseAgain()

	sent, err := repo.ProcessParticipant(context.Background(), participant, func(current notificationmodel.Participant, delivered map[int]bool) (*int, error) {
		if current.Email != participant.Email || delivered[25] {
			t.Fatalf("current=%+v delivered=%v", current, delivered)
		}
		threshold := 25
		return &threshold, nil
	})
	if err != nil || !sent {
		t.Fatalf("first delivery sent=%v error=%v", sent, err)
	}

	sent, err = repo.ProcessParticipant(context.Background(), participant, func(_ notificationmodel.Participant, delivered map[int]bool) (*int, error) {
		if !delivered[25] {
			t.Fatal("recorded threshold is missing")
		}
		return nil, nil
	})
	if err != nil || sent {
		t.Fatalf("deduplicated delivery sent=%v error=%v", sent, err)
	}
}

func notificationTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("GAME_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("GAME_TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Apply(context.Background(), db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func insertNotificationParticipant(t *testing.T, db *sql.DB) notificationmodel.Participant {
	t.Helper()
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	participant := notificationmodel.Participant{UserID: uuid.New(), Email: fmt.Sprintf("notification-%s@example.com", uuid.NewString()), EnergyPercent: 25, EnergyUpdatedAt: now}
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO users (id, email, display_name, password_hash, created_at)
		VALUES ($1, $2, 'Notification_Test', 'test-hash', $3)
	`, participant.UserID, participant.Email, now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, participant.UserID)
	})
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO pets (id, user_id, xp, energy_percent, energy_updated_at, last_charged_at, created_at, updated_at)
		VALUES ($1, $2, 0, $3, $4, $4, $4, $4)
	`, uuid.New(), participant.UserID, participant.EnergyPercent, participant.EnergyUpdatedAt)
	if err != nil {
		t.Fatalf("insert pet: %v", err)
	}
	return participant
}
