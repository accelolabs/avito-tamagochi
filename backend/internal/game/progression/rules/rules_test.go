package rules

import (
	"testing"
	"time"
)

func TestLevelFromXP(t *testing.T) {
	for _, test := range []struct{ xp, level int }{{0, 1}, {99, 1}, {100, 2}, {200, 3}} {
		if got := LevelFromXP(test.xp); got != test.level {
			t.Errorf("LevelFromXP(%d) = %d, want %d", test.xp, got, test.level)
		}
	}
}

func TestEnergyPercent(t *testing.T) {
	chargedAt := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name string
		now  time.Time
		want int
	}{
		{"full", chargedAt, 100}, {"half", chargedAt.Add(24 * time.Hour), 50},
		{"empty", chargedAt.Add(48 * time.Hour), 0}, {"after empty", chargedAt.Add(72 * time.Hour), 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := EnergyPercent(chargedAt, test.now); got != test.want {
				t.Errorf("EnergyPercent() = %d, want %d", got, test.want)
			}
		})
	}
}
