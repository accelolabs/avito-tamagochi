package errors

import "errors"

var (
	ErrInvalidAction     = errors.New("invalid game action")
	ErrChargeAlreadyUsed = errors.New("charge already used today")
	ErrNotImplemented    = errors.New("domain service is not implemented")
)
