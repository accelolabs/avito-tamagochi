package service

import (
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/pet"
)

func TestStats(t *testing.T) {
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	value := &pet.Pet{
		XP:            100,
		LastChargedAt: now.Add(-24 * time.Hour),
	}

	got := stats(value, now)
	if got.XP != 100 || got.Level != 2 || got.Energy != 50 {
		t.Fatalf("stats() = %+v, want XP=100 level=2 energy=50", got)
	}
}
