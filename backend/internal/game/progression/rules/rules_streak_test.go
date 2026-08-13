package rules

import (
	"testing"
	"time"
)

func TestIsDead(t *testing.T) {
	tests := []struct {
		name   string
		energy int
		want   bool
	}{
		{"charged", 100, false},
		{"almost empty", 1, false},
		{"empty", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDead(tt.energy)
			if got != tt.want {
				t.Errorf("IsDead(%d) = %v, want %v", tt.energy, got, tt.want)
			}
		})
	}
}

func TestEnergyDecayDuration(t *testing.T) {
	if EnergyDecayDuration != 48*time.Hour {
		t.Errorf("EnergyDecayDuration = %v, want 48h", EnergyDecayDuration)
	}
}

func TestDailyRewardXP(t *testing.T) {
	tests := []struct {
		streak int
		want   int
	}{
		{streak: 0, want: 10},
		{streak: 1, want: 10},
		{streak: 2, want: 15},
		{streak: 8, want: 45},
		{streak: 9, want: 50},
		{streak: 100, want: 50},
	}
	for _, test := range tests {
		if got := DailyRewardXP(test.streak); got != test.want {
			t.Errorf("DailyRewardXP(%d) = %d, want %d", test.streak, got, test.want)
		}
	}
}

func TestCurrentStreakUsesMoscowCalendarDays(t *testing.T) {
	today := time.Date(2026, 8, 13, 0, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	if got := CurrentStreak(4, &today, today); got != 4 {
		t.Errorf("today streak = %d, want 4", got)
	}
	if got := CurrentStreak(4, &yesterday, today); got != 4 {
		t.Errorf("yesterday streak = %d, want 4", got)
	}
	if got := CurrentStreak(4, &twoDaysAgo, today); got != 0 {
		t.Errorf("expired streak = %d, want 0", got)
	}
	if got := CurrentStreak(4, nil, today); got != 0 {
		t.Errorf("streak without date = %d, want 0", got)
	}
}
