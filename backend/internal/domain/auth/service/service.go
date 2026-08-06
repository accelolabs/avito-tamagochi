package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.User, *auth.Session, error)
	Login(ctx context.Context, req auth.LoginRequest) (*auth.User, *auth.Session, error)
	Logout(ctx context.Context, sessionID string) error
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.User, *auth.Session, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := auth.User{
		ID:           uuid.NewString(),
		Email:        normalizedEmail,
		DisplayName:  req.DisplayName,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}

	pet := auth.Pet{
		ID:                uuid.NewString(),
		Name:              "Кита (K1-T4)",
		Level:             1,
		TotalXP:           0,
		NextLevelXP:       100,
		Stage:             "egg",
		BatteryLevel:      100,
		Status:            "happy",
		IsActionAvailable: true,
		UpdatedAt:         now,
	}

	session := auth.Session{
		ID:        uuid.NewString(),
		ExpiresAt: now.Add(604800 * time.Second), // 7 days
		CreatedAt: now,
	}

	createdUser, createdSession, err := s.repo.CreateUserAndPet(ctx, user, pet, session)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, nil, ErrEmailAlreadyExists
		}
		return nil, nil, fmt.Errorf("failed to create user and pet: %w", err)
	}

	return createdUser, createdSession, nil
}

func (s *authService) Login(ctx context.Context, req auth.LoginRequest) (*auth.User, *auth.Session, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.repo.FindUserByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	ok, err := s.verifyPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		return nil, nil, ErrInvalidCredentials
	}

	now := time.Now()
	session := auth.Session{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		ExpiresAt: now.Add(604800 * time.Second), // 7 days
		CreatedAt: now,
	}

	createdSession, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return user, createdSession, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

func (s *authService) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 64*1024, 1, 4, b64Salt, b64Hash)
	return encodedHash, nil
}

func (s *authService) verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	var m, t, p uint32
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("incompatible argon2 version")
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}
