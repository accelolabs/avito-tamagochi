package service

import (
	"context"
	"testing"
	"time"

	leadermodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/model"
	"github.com/google/uuid"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type recordingRepository struct {
	deltaSince  []time.Time
	streakSince []time.Time
}

func (r *recordingRepository) GetTopByXP(context.Context) ([]leadermodel.XPEntry, error) {
	return []leadermodel.XPEntry{}, nil
}

func (r *recordingRepository) GetUserRankByXP(context.Context, uuid.UUID) (*leadermodel.XPEntry, error) {
	return nil, nil
}

func (r *recordingRepository) GetTopByXPDelta(_ context.Context, since time.Time) ([]leadermodel.XPEntry, error) {
	r.deltaSince = append(r.deltaSince, since)
	return []leadermodel.XPEntry{}, nil
}

func (r *recordingRepository) GetUserRankByXPDelta(_ context.Context, _ uuid.UUID, since time.Time) (*leadermodel.XPEntry, error) {
	r.deltaSince = append(r.deltaSince, since)
	return nil, nil
}

func (r *recordingRepository) GetTopByStreak(_ context.Context, since time.Time) ([]leadermodel.StreakEntry, error) {
	r.streakSince = append(r.streakSince, since)
	return []leadermodel.StreakEntry{}, nil
}

func (r *recordingRepository) GetUserRankByStreak(_ context.Context, _ uuid.UUID, since time.Time) (*leadermodel.StreakEntry, error) {
	r.streakSince = append(r.streakSince, since)
	return nil, nil
}

func TestDeltaLeaderboardsUseRollingWindows(t *testing.T) {
	now := time.Date(2026, 8, 13, 18, 30, 0, 0, time.FixedZone("test", 5*60*60))
	repo := &recordingRepository{}
	service := NewWithClock(repo, fixedClock{now: now})

	if _, err := service.GetWeekly(context.Background(), uuid.New()); err != nil {
		t.Fatalf("get weekly leaderboard: %v", err)
	}
	if _, err := service.GetMonthly(context.Background(), uuid.New()); err != nil {
		t.Fatalf("get monthly leaderboard: %v", err)
	}

	want := []time.Time{
		now.UTC().Add(-7 * 24 * time.Hour), now.UTC().Add(-7 * 24 * time.Hour),
		now.UTC().Add(-30 * 24 * time.Hour), now.UTC().Add(-30 * 24 * time.Hour),
	}
	if len(repo.deltaSince) != len(want) {
		t.Fatalf("recorded %d boundaries, want %d", len(repo.deltaSince), len(want))
	}
	for i := range want {
		if !repo.deltaSince[i].Equal(want[i]) {
			t.Errorf("boundary %d = %v, want %v", i, repo.deltaSince[i], want[i])
		}
	}
}

func TestStreakLeaderboardKeepsYesterdayActive(t *testing.T) {
	now := time.Date(2026, 8, 13, 23, 30, 0, 0, time.UTC)
	repo := &recordingRepository{}
	service := NewWithClock(repo, fixedClock{now: now})

	if _, err := service.GetStreak(context.Background(), uuid.New()); err != nil {
		t.Fatalf("get streak leaderboard: %v", err)
	}

	want := time.Date(2026, 8, 13, 0, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
	for i, value := range repo.streakSince {
		if !value.Equal(want) {
			t.Errorf("boundary %d = %v, want %v", i, value, want)
		}
	}
}
