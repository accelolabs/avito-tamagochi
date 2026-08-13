package model

import (
	"time"

	"github.com/google/uuid"
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
	Password    string `json:"-"`
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string `json:"-"`
}
