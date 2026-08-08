package progression

import (
	"testing"
	"time"
)

func TestLevelFromXP(t *testing.T) {
	tests := []struct {
		xp    int
		level int
	}{
		{xp: 0, level: 1},
		{xp: 99, level: 1},
		{xp: 100, level: 2},
		{xp: 200, level: 3},
	}

	for _, test := range tests {
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
		{name: "full", now: chargedAt, want: 100},
		{name: "half", now: chargedAt.Add(24 * time.Hour), want: 50},
		{name: "empty", now: chargedAt.Add(48 * time.Hour), want: 0},
		{name: "after empty", now: chargedAt.Add(72 * time.Hour), want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := EnergyPercent(chargedAt, test.now); got != test.want {
				t.Errorf("EnergyPercent() = %d, want %d", got, test.want)
			}
		})
	}
}
