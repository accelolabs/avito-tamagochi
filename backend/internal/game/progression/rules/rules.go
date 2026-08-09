package rules

import (
	"time"

	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
)

const ChargeXPAmount = 10

func LevelFromXP(xp int) int {
	if xp < 0 {
		xp = 0
	}
	return xp/100 + 1
}

func StageFromLevel(level int) petmodel.Stage {
	switch {
	case level >= 9:
		return petmodel.Adult
	case level >= 6:
		return petmodel.Teen
	case level >= 3:
		return petmodel.Child
	default:
		return petmodel.Egg
	}
}

func EnergyPercent(lastChargedAt, now time.Time) int {
	energy := 100 - int(now.Sub(lastChargedAt).Seconds()/(48*time.Hour).Seconds()*100)
	if energy < 0 {
		return 0
	}
	if energy > 100 {
		return 100
	}
	return energy
}

func TaskXP(taskType taskmodel.Type) int {
	switch taskType {
	case taskmodel.View:
		return 20
	case taskmodel.Favorite:
		return 25
	case taskmodel.CreateListing:
		return 40
	case taskmodel.CreateListingCategory:
		return 50
	default:
		return 0
	}
}

func RewardTypeForLevel(level int) rewardmodel.Type {
	if level < 2 {
		return ""
	}
	if level%2 == 0 {
		return rewardmodel.Promotion
	}
	return rewardmodel.Delivery
}
