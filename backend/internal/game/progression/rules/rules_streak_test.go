package rules

import (
	"testing"
	"time"
)

func TestIsDead(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		lastChargedAt time.Time
		want          bool
	}{
		{"just charged", now, false},
		{"charged 1 hour ago", now.Add(-1 * time.Hour), false},
		{"charged 47h59m ago", now.Add(-47*time.Hour - 59*time.Minute), false},
		{"charged exactly 48h ago", now.Add(-48 * time.Hour), true},
		{"charged 49h ago", now.Add(-49 * time.Hour), true},
		{"charged 72h ago", now.Add(-72 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDead(tt.lastChargedAt, now)
			if got != tt.want {
				t.Errorf("IsDead(%v, %v) = %v, want %v", tt.lastChargedAt, now, got, tt.want)
			}
		})
	}
}

func TestEnergyDecayDuration(t *testing.T) {
	if EnergyDecayDuration != 48*time.Hour {
		t.Errorf("EnergyDecayDuration = %v, want 48h", EnergyDecayDuration)
	}
}
