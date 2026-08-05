package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, *auth.Session, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, *auth.Session, error) {
	// Basic validation
	if req.Email == "" {
		return nil, nil, ErrInvalidEmail
	}
	if len(req.Password) < 8 {
		return nil, nil, ErrInvalidPassword
	}

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

func (s *authService) hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format is $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, 64*1024, 1, 4, b64Salt, b64Hash)

	return encodedHash, nil
}
