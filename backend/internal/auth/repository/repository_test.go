package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	autherrors "github.com/accelolabs/avito-tamagochi/backend/internal/auth/errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func testDatabase(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("AUTH_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("AUTH_TEST_DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Skipf("open test database: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Skipf("connect to test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testUser() model.User {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return model.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("auth-test-%s@example.com", uuid.NewString()),
		DisplayName:  "Test_Player",
		PasswordHash: "test-password-hash",
		CreatedAt:    now,
	}
}

func insertTestUser(t *testing.T, db *sql.DB, user model.User) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `INSERT INTO users (id, email, display_name, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID) })
}

func TestRepositoryFindUserByEmailAndID(t *testing.T) {
	db := testDatabase(t)
	repository := New(db)
	user := testUser()
	insertTestUser(t, db, user)

	byEmail, err := repository.FindUserByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("FindUserByEmail() error = %v", err)
	}
	if *byEmail != user {
		t.Fatalf("FindUserByEmail() = %+v, want %+v", *byEmail, user)
	}

	byID, err := repository.FindUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v", err)
	}
	if *byID != user {
		t.Fatalf("FindUserByID() = %+v, want %+v", *byID, user)
	}
}

func TestRepositoryFindSession(t *testing.T) {
	db := testDatabase(t)
	repository := New(db)
	user := testUser()
	insertTestUser(t, db, user)
	session := model.Session{ID: uuid.New(), UserID: user.ID, CreatedAt: time.Now().UTC().Truncate(time.Microsecond), ExpiresAt: time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)}
	_, err := db.ExecContext(context.Background(), `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4)`, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		t.Fatalf("insert test session: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM sessions WHERE id = $1`, session.ID) })

	got, err := repository.FindSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("FindSession() error = %v", err)
	}
	if *got != session {
		t.Fatalf("FindSession() = %+v, want %+v", *got, session)
	}
}

func TestCreateUserMapsUniqueViolation(t *testing.T) {
	db := testDatabase(t)
	repository := New(db)
	existing := testUser()
	insertTestUser(t, db, existing)
	duplicate := testUser()
	duplicate.Email = existing.Email

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	err = repository.CreateUser(context.Background(), tx, duplicate)
	_ = tx.Rollback()
	if !errors.Is(err, autherrors.ErrEmailAlreadyExists) {
		t.Fatalf("CreateUser() error = %v, want %v", err, autherrors.ErrEmailAlreadyExists)
	}
}

func TestIsUniqueViolationAcceptsWrappedPostgresError(t *testing.T) {
	err := fmt.Errorf("insert user: %w", &pq.Error{Code: uniqueViolationCode})
	if !isUniqueViolation(err) {
		t.Fatal("isUniqueViolation() = false for wrapped unique violation")
	}
}

func TestRepositoryCreateSessionReturnsDatabaseError(t *testing.T) {
	db := testDatabase(t)
	repository := New(db)
	session := model.Session{ID: uuid.New(), UserID: uuid.New(), CreatedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(time.Hour)}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	err = repository.CreateSession(context.Background(), tx, session)
	_ = tx.Rollback()
	if err == nil {
		t.Fatal("CreateSession() error = nil, want foreign key error")
	}
}

func TestRepositoryDeleteSessionIsIdempotent(t *testing.T) {
	db := testDatabase(t)
	repository := New(db)
	if err := repository.DeleteSession(context.Background(), uuid.New()); err != nil {
		t.Fatalf("DeleteSession() error = %v for missing session", err)
	}
}
