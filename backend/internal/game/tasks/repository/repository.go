package repository

import (
	"context"
	"database/sql"
	"time"

	taskmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/model"
	"github.com/google/uuid"
)

type Repository interface {
	GetTodayProgress(context.Context, uuid.UUID, time.Time) ([]taskmodel.Progress, error)
	GetProgressForUpdate(context.Context, *sql.Tx, uuid.UUID, time.Time, taskmodel.Type) (*taskmodel.Progress, error)
	SaveProgress(context.Context, *sql.Tx, taskmodel.Progress) error
}
