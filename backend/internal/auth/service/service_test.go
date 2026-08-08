package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/security"
	"github.com/google/uuid"
)

type serviceRepository struct {
	deleteSession func(context.Context, uuid.UUID) error
}

func (r serviceRepository) FindUserByEmail(context.Context, string) (*User, error) {
	return nil, sql.ErrNoRows
}
func (r serviceRepository) FindUserByID(context.Context, uuid.UUID) (*User, error) {
	return nil, sql.ErrNoRows
}
func (r serviceRepository) CreateUser(context.Context, *sql.Tx, User) error       { return nil }
func (r serviceRepository) CreateSession(context.Context, *sql.Tx, Session) error { return nil }
func (r serviceRepository) FindSession(context.Context, uuid.UUID) (*Session, error) {
	return nil, sql.ErrNoRows
}
func (r serviceRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return r.deleteSession(ctx, id)
}

func TestLogoutReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("delete session failed")
	service := NewService(nil, serviceRepository{deleteSession: func(context.Context, uuid.UUID) error { return wantErr }})

	if err := service.Logout(context.Background(), uuid.New()); !errors.Is(err, wantErr) {
		t.Fatalf("Logout() error = %v, want %v", err, wantErr)
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  USER@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeEmail() = %q", got)
	}
}

func TestValidateRegistration(t *testing.T) {
	valid := []struct {
		name    string
		email   string
		pass    string
		display string
	}{
		{name: "valid", email: "user@example.com", pass: "password", display: "Player_One"},
	}
	for _, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			if err := validateRegistration(test.email, test.pass, test.display); err != nil {
				t.Fatalf("validateRegistration() error = %v", err)
			}
		})
	}

	for _, test := range []struct {
		name    string
		email   string
		pass    string
		display string
	}{
		{name: "short password", email: "user@example.com", pass: "short", display: "Player"},
		{name: "invalid email", email: "not-an-email", pass: "password", display: "Player"},
		{name: "invalid display name", email: "user@example.com", pass: "password", display: "Игрок"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validateRegistration(test.email, test.pass, test.display); err != ErrInvalidInput {
				t.Fatalf("validateRegistration() error = %v, want %v", err, ErrInvalidInput)
			}
		})
	}
}

func TestPasswordHash(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword() error = %v", err)
	}
	if strings.Contains(hash, password) {
		t.Fatal("password hash contains plaintext password")
	}
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Fatalf("hash does not use Argon2id PHC format: %q", hash)
	}
	if !security.VerifyPassword(password, hash) {
		t.Fatal("verifyPassword() rejected the original password")
	}
	if security.VerifyPassword("wrong password", hash) {
		t.Fatal("verifyPassword() accepted a wrong password")
	}
}
