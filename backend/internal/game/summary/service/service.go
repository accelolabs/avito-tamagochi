package service

import (
	"context"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	summarymodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/model"
	"github.com/google/uuid"
)

type Service interface {
	GetToday(context.Context, uuid.UUID) (*summarymodel.DailySummary, error)
}

type service struct{}

func New() Service { return service{} }
func (service) GetToday(context.Context, uuid.UUID) (*summarymodel.DailySummary, error) {
	return nil, gameerrors.ErrNotImplemented
}
