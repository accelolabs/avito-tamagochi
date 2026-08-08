package auth

import (
	"strings"
	"testing"
)

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
	hash := hashPassword(password)
	if strings.Contains(hash, password) {
		t.Fatal("password hash contains plaintext password")
	}
	if !verifyPassword(password, hash) {
		t.Fatal("verifyPassword() rejected the original password")
	}
	if verifyPassword("wrong password", hash) {
		t.Fatal("verifyPassword() accepted a wrong password")
	}
}
