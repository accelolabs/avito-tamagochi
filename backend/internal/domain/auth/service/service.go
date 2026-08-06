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
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, *auth.Session, error)
	Login(ctx context.Context, req auth.LoginRequest) (*auth.Session, error)
	Logout(ctx context.Context, sessionID string) error
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, *auth.Session, error) {
	// Validation is now handled by Gin's binding
	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := auth.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: passwordHash,
		CreatedAt:    now,
	}
	pet := auth.Pet{
		ID:                uuid.NewString(),
		Name:              req.PetName,
		Level:             1,
		XP:                0,
		XPToNextLevel:     100,
		BatteryLevel:      100,
		Status:            "happy",
		IsActionAvailable: true,
		UpdatedAt:         now,
	}
	session := auth.Session{
		ID:        uuid.NewString(),
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}

	createdUser, createdPet, createdSession, err := s.repo.CreateUserAndPet(ctx, user, pet, session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user and pet: %w", err)
	}

	resp := &auth.AuthResponse{
		User: auth.User{
			ID:          createdUser.ID,
			Email:       createdUser.Email,
			DisplayName: createdUser.DisplayName,
			CreatedAt:   createdUser.CreatedAt,
		},
		Pet: *createdPet,
	}

	return resp, createdSession, nil
}

func (s *authService) Login(ctx context.Context, req auth.LoginRequest) (*auth.Session, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := s.verifyPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	session := auth.Session{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
	}

	return s.repo.CreateSession(ctx, session)
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	return s.repo.DeleteSession(ctx, sessionID)
}

// --- Password Hashing ---

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

	var version, m, t, p uint32
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
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

	otherHash := argon2.IDKey([]byte(password), salt, uint32(t), uint32(m), uint32(p), uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}
