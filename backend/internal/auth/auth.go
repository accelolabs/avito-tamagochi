package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidInput       = errors.New("invalid input")
)

type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type Repository interface {
	FindUserByEmail(context.Context, string) (*User, error)
	FindUserByID(context.Context, uuid.UUID) (*User, error)
	CreateUser(context.Context, *sql.Tx, User) error
	CreateSession(context.Context, *sql.Tx, Session) error
	FindSession(context.Context, uuid.UUID) (*Session, error)
	DeleteSession(context.Context, uuid.UUID) error
}

type Service interface {
	Register(context.Context, RegisterInput) (*User, *Session, error)
	Login(context.Context, LoginInput) (*User, *Session, error)
	Logout(context.Context, uuid.UUID) error
	ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error)
	FindUser(context.Context, uuid.UUID) (*User, error)
}
