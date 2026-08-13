package errors

import "errors"

var (
	ErrInvalidAction     = errors.New("invalid game action")
	ErrChargeAlreadyUsed = errors.New("charge already used today")
	ErrTaskNotAvailable  = errors.New("task is not available today")
	ErrPetDead           = errors.New("pet is dead")
	ErrRewardNotFound    = errors.New("reward not found")
	ErrRewardAlreadyUsed = errors.New("reward already used")
)
