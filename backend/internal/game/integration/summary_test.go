package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	summaryrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/repository"
	"github.com/google/uuid"
)

func TestSummaryCountsCompletedTasksInsteadOfTaskXPEvents(t *testing.T) {
	db := testDatabase(t)
	userID := insertTestUser(t, db)
	localDate := clock.MoscowDate(rollbackTestTime)
	petID := insertPet(t, db, userID, 60, rollbackTestTime.Add(-24*time.Hour))

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO task_progress (user_id, local_date, task_type, progress, completed_at)
		VALUES ($1, $2, 'view', 5, $3)
	`, userID, localDate.Format("2006-01-02"), rollbackTestTime)
	if err != nil {
		t.Fatalf("insert completed task: %v", err)
	}

	for index := 1; index <= 2; index++ {
		_, err := db.ExecContext(context.Background(), `
			INSERT INTO xp_events (id, user_id, pet_id, source, source_key, amount, occurred_at, local_date)
			VALUES ($1, $2, $3, 'task', $4, 10, $5, $6)
		`, uuid.New(), userID, petID, uuid.NewString(), rollbackTestTime, localDate.Format("2006-01-02"))
		if err != nil {
			t.Fatalf("insert task XP event %d: %v", index, err)
		}
	}

	value, err := summaryrepository.New(db).GetToday(context.Background(), userID, localDate, rollbackTestTime)
	if err != nil {
		t.Fatalf("GetToday() error = %v", err)
	}
	if value.CompletedTasks != 1 {
		t.Fatalf("CompletedTasks = %d, want 1", value.CompletedTasks)
	}
	if value.XPEarned != 20 {
		t.Fatalf("XPEarned = %d, want 20", value.XPEarned)
	}
}
