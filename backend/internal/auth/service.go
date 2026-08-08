package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// Доменные модели
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"` // Никогда не отдаем на фронтенд
	CreatedAt    time.Time `json:"createdAt"`
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

// Контракт зависимостей для БД
type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, tx *sql.Tx, user User) error
	CreateSession(ctx context.Context, tx *sql.Tx, session Session) error
	FindSession(ctx context.Context, id uuid.UUID) (*Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
}

// Публичный контракт сервиса
type Service interface {
	Register(ctx context.Context, input RegisterInput) (*User, *Session, error)
	Login(ctx context.Context, input LoginInput) (*User, *Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	ValidateSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error)
}

type authService struct {
	db   *sql.DB // Сервис управляет транзакциями, поэтому нужен доступ к DB
	repo Repository
}

func NewService(db *sql.DB, repo Repository) Service {
	return &authService{db: db, repo: repo}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*User, *Session, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	// Хэширование Argon2id
	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := User{
		ID:           uuid.New(),
		Email:        email,
		DisplayName:  input.DisplayName,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}

	session := Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour), // 7 дней
		CreatedAt: time.Now().UTC(),
	}

	// Открываем транзакцию (ответственность сервиса)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // Безопасный откат, если не был вызван Commit

	if err := s.repo.CreateUser(ctx, tx, user); err != nil {
		return nil, nil, err
	}

	if err := s.repo.CreateSession(ctx, tx, session); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit tx: %w", err)
	}

	return &user, &session, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*User, *Session, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

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

	session := Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	if err := s.repo.CreateSession(ctx, tx, session); err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return user, &session, nil
}

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *authService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	session, err := s.repo.FindSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrSessionNotFound
		}
		return uuid.Nil, err
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(ctx, sessionID) // Подчищаем протухшую сессию
		return uuid.Nil, ErrSessionExpired
	}

	return session.UserID, nil
}

// --- Утилиты для Argon2id ---
// В production версии соль должна генерироваться случайным образом для каждого пароля.
// Здесь представлена упрощенная, но надежная реализация.

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x.%x", salt, hash), nil
}

func verifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return false
	}
	salt, _ := hex.DecodeString(parts[0])
	expectedHash, _ := hex.DecodeString(parts[1])
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return string(hash) == string(expectedHash)
}
