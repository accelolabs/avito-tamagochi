package game

import "errors"

var (
	ErrInvalidAction     = errors.New("invalid game action")
	ErrChargeAlreadyUsed = errors.New("charge already used today")
)
