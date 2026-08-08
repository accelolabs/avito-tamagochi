package service

import (
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
)

type Service interface {
	LevelFromXP(int) int
	EnergyPercent(time.Time, time.Time) int
	MoscowDate(time.Time) time.Time
	ChargeXP() int
}

type service struct{}

func New() Service                                    { return service{} }
func (service) LevelFromXP(xp int) int                { return rules.LevelFromXP(xp) }
func (service) EnergyPercent(last, now time.Time) int { return rules.EnergyPercent(last, now) }
func (service) MoscowDate(now time.Time) time.Time    { return rules.MoscowDate(now) }
func (service) ChargeXP() int                         { return rules.ChargeXPAmount }
