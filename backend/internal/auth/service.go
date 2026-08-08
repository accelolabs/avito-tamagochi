package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/argon2"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidInput       = errors.New("invalid input")
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
	minDisplayNameLen = 2
	maxDisplayNameLen = 32
	sessionLifetime   = 7 * 24 * time.Hour
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

	hash := hashPassword(input.Password)
	now := s.now().UTC()
	user := User{ID: uuid.New(), Email: email, DisplayName: input.DisplayName, PasswordHash: hash, CreatedAt: now}
	session := newSession(user.ID, now)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.CreateUser(ctx, tx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, nil, ErrEmailAlreadyExists
		}
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
	if email == "" || len(input.Password) < minPasswordLength || len(input.Password) > maxPasswordLength {
		return nil, nil, ErrInvalidInput
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}
	if !verifyPassword(input.Password, user.PasswordHash) {
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
	err := s.repo.DeleteSession(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
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
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return ErrInvalidInput
	}
	if len(displayName) < minDisplayNameLen || len(displayName) > maxDisplayNameLen {
		return ErrInvalidInput
	}
	for _, char := range displayName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || char == '-') {
			return ErrInvalidInput
		}
	}
	return nil
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic("crypto/rand failed")
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	encoded := append(append([]byte{}, salt...), hash...)
	return base64.RawStdEncoding.EncodeToString(encoded)
}

func verifyPassword(password, encoded string) bool {
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) != 48 {
		return false
	}
	hash := argon2.IDKey([]byte(password), decoded[:16], 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, decoded[16:]) == 1
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
