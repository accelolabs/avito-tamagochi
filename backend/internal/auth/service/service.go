package service

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"time"

	autherrors "github.com/accelolabs/avito-tamagochi/backend/internal/auth/errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	authrepository "github.com/accelolabs/avito-tamagochi/backend/internal/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/security"
	"github.com/google/uuid"
)

const (
	minDisplayNameLen = 2
	maxDisplayNameLen = 32
	sessionLifetime   = 7 * 24 * time.Hour
)

type User = model.User
type Session = model.Session
type RegisterInput = model.RegisterInput
type LoginInput = model.LoginInput
type Repository = authrepository.Repository

type Service interface {
	Register(context.Context, RegisterInput) (*User, *Session, error)
	Login(context.Context, LoginInput) (*User, *Session, error)
	Logout(context.Context, uuid.UUID) error
	ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error)
	FindUser(context.Context, uuid.UUID) (*User, error)
}

var (
	ErrEmailAlreadyExists = autherrors.ErrEmailAlreadyExists
	ErrInvalidCredentials = autherrors.ErrInvalidCredentials
	ErrSessionNotFound    = autherrors.ErrSessionNotFound
	ErrInvalidInput       = autherrors.ErrInvalidInput
)

type service struct {
	db   *sql.DB
	repo Repository
	now  func() time.Time
}

func NewService(db *sql.DB, repo Repository) Service {
	return &service{db: db, repo: repo, now: time.Now}
}

func (s *service) Register(ctx context.Context, input RegisterInput) (*User, *Session, error) {
	email := normalizeEmail(input.Email)
	if err := validateRegistration(email, input.Password, input.DisplayName); err != nil {
		return nil, nil, err
	}

	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, nil, err
	}
	now := s.now().UTC()
	user := User{ID: uuid.New(), Email: email, DisplayName: input.DisplayName, PasswordHash: hash, CreatedAt: now}
	session := newSession(user.ID, now)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.CreateUser(ctx, tx, user); err != nil {
		return nil, nil, err
	}
	if err := s.repo.CreateSession(ctx, tx, session); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return &user, &session, nil
}

func (s *service) Login(ctx context.Context, input LoginInput) (*User, *Session, error) {
	email := normalizeEmail(input.Email)
	if email == "" || !security.ValidPassword(input.Password) {
		return nil, nil, ErrInvalidInput
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if !security.VerifyPassword(input.Password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	now := s.now().UTC()
	session := newSession(user.ID, now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.repo.CreateSession(ctx, tx, session); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return user, &session, nil
}

func (s *service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *service) ValidateSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	session, err := s.repo.FindSession(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	if !s.now().Before(session.ExpiresAt) {
		return uuid.Nil, ErrSessionNotFound
	}
	return session.UserID, nil
}

func (s *service) FindUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func newSession(userID uuid.UUID, now time.Time) Session {
	return Session{ID: uuid.New(), UserID: userID, CreatedAt: now, ExpiresAt: now.Add(sessionLifetime)}
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func validateRegistration(email, password, displayName string) error {
	if email == "" || len(email) > 254 {
		return ErrInvalidInput
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return ErrInvalidInput
	}
	if !security.ValidPassword(password) {
		return ErrInvalidInput
	}
	if len(displayName) < minDisplayNameLen || len(displayName) > maxDisplayNameLen {
		return ErrInvalidInput
	}
	for _, char := range displayName {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && char != '_' && char != '-' {
			return ErrInvalidInput
		}
	}
	return nil
}
