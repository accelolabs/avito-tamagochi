package tasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetTodayProgress(context.Context, uuid.UUID, time.Time) ([]TaskProgress, error)
	GetProgressForUpdate(context.Context, *sql.Tx, uuid.UUID, time.Time, TaskType) (*TaskProgress, error)
	SaveProgress(context.Context, *sql.Tx, TaskProgress) error
}
