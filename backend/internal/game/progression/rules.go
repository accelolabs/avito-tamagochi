package progression

import (
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game"
)

const ChargeXPAmount = 10

func LevelFromXP(xp int) int {
	if xp < 0 {
		xp = 0
	}
	return xp/100 + 1
}

func EnergyPercent(lastChargedAt, now time.Time) int {
	elapsed := now.Sub(lastChargedAt)
	energy := 100 - int(elapsed.Seconds()/(48*time.Hour).Seconds()*100)
	if energy < 0 {
		return 0
	}
	if energy > 100 {
		return 100
	}
	return energy
}

func MoscowDate(now time.Time) time.Time {
	return game.MoscowDate(now)
}
