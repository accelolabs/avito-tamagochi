package service

import (
	"context"

	gameerrors "github.com/accelolabs/avito-tamagochi/backend/internal/game/errors"
	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	taskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	"github.com/google/uuid"
)

type Service interface {
	GetTodayTasks(context.Context, uuid.UUID) ([]taskmodel.Progress, error)
	ApplyAction(context.Context, uuid.UUID, taskmodel.Type) error
}

type service struct{ repo taskrepository.Repository }

func New(repo taskrepository.Repository) Service { return &service{repo: repo} }
func (s *service) GetTodayTasks(context.Context, uuid.UUID) ([]taskmodel.Progress, error) {
	return nil, gameerrors.ErrNotImplemented
}
func (s *service) ApplyAction(context.Context, uuid.UUID, taskmodel.Type) error {
	return gameerrors.ErrNotImplemented
}
