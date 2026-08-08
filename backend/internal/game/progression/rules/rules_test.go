package rules

import (
	"github.com/accelolabs/avito-tamagochi/backend/internal/game/clock"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
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

func TestRewardTypeForLevel(t *testing.T) {
	for level := 1; level <= 10; level++ {
		want := rewardmodel.Delivery
		if level < 2 {
			want = ""
		}
		if level >= 2 && level%2 == 0 {
			want = rewardmodel.Promotion
		}
		if got := RewardTypeForLevel(level); got != want {
			t.Errorf("level %d: got %q, want %q", level, got, want)
		}
	}
}

func TestTaskXP(t *testing.T) {
	for task, want := range map[taskmodel.Type]int{
		taskmodel.View: 20, taskmodel.Favorite: 25, taskmodel.CreateListing: 40, taskmodel.CreateListingCategory: 50,
	} {
		if got := TaskXP(task); got != want {
			t.Errorf("TaskXP(%q) = %d, want %d", task, got, want)
		}
	}
}

func TestMoscowDate(t *testing.T) {
	before := time.Date(2026, 8, 9, 20, 59, 0, 0, time.UTC)
	after := before.Add(2 * time.Minute)
	if clock.MoscowDate(before).Day() != 9 || clock.MoscowDate(after).Day() != 10 {
		t.Fatal("MoscowDate did not cross midnight at the Moscow boundary")
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
