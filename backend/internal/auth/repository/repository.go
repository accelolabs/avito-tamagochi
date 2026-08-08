package repository

import (
	"context"
	"database/sql"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	"github.com/google/uuid"
)

type Repository interface {
	FindUserByEmail(context.Context, string) (*model.User, error)
	FindUserByID(context.Context, uuid.UUID) (*model.User, error)
	CreateUser(context.Context, *sql.Tx, model.User) error
	CreateSession(context.Context, *sql.Tx, model.Session) error
	FindSession(context.Context, uuid.UUID) (*model.Session, error)
	DeleteSession(context.Context, uuid.UUID) error
}
