package service

import (
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/rules"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
)

type Service interface {
	LevelFromXP(int) int
	EnergyPercent(time.Time, time.Time) int
	MoscowDate(time.Time) time.Time
	ChargeXP() int
	TaskXP(taskmodel.Type) int
	RewardTypeForLevel(int) rewardmodel.Type
}

type service struct{}

func New() Service                                    { return service{} }
func (service) LevelFromXP(xp int) int                { return rules.LevelFromXP(xp) }
func (service) EnergyPercent(last, now time.Time) int { return rules.EnergyPercent(last, now) }
func (service) MoscowDate(now time.Time) time.Time    { return rules.MoscowDate(now) }
func (service) ChargeXP() int                         { return rules.ChargeXPAmount }
func (service) TaskXP(taskType taskmodel.Type) int    { return rules.TaskXP(taskType) }
func (service) RewardTypeForLevel(level int) rewardmodel.Type {
	return rules.RewardTypeForLevel(level)
}
