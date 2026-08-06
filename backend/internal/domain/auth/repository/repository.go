package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

const pgUniqueViolationCode = "23505"

type AuthRepository interface {
	CreateUserAndPet(ctx context.Context, user auth.User, pet auth.Pet, session auth.Session) (*auth.User, *auth.Session, error)
	FindUserByEmail(ctx context.Context, email string) (*auth.User, error)
	CreateSession(ctx context.Context, session auth.Session) (*auth.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type pgRepository struct {
	db *sql.DB
}

func NewPgRepository(db *sql.DB) AuthRepository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateUserAndPet(ctx context.Context, user auth.User, pet auth.Pet, session auth.Session) (*auth.User, *auth.Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userQuery := `INSERT INTO users (id, email, display_name, password_hash, created_at) 
				   VALUES ($1, $2, $3, $4, $5) RETURNING id, email, display_name, created_at`
	err = tx.QueryRowContext(ctx, userQuery, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == pgUniqueViolationCode {
			return nil, nil, ErrEmailAlreadyExists
		}
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	pet.OwnerID = user.ID
	petQuery := `INSERT INTO pets (id, owner_id, name, level, total_xp, next_level_xp, stage, battery_level, status, is_action_available, updated_at) 
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.ExecContext(ctx, petQuery, pet.ID, pet.OwnerID, pet.Name, pet.Level, pet.TotalXP, pet.NextLevelXP, pet.Stage, pet.BatteryLevel, pet.Status, pet.IsActionAvailable, pet.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pet: %w", err)
	}

	session.UserID = user.ID
	sessionQuery := `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4) RETURNING id, user_id, expires_at, created_at`
	err = tx.QueryRowContext(ctx, sessionQuery, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, &session, nil
}

func (r *pgRepository) FindUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	user := &auth.User{}
	query := `SELECT id, email, display_name, password_hash, created_at FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

func (r *pgRepository) CreateSession(ctx context.Context, session auth.Session) (*auth.Session, error) {
	query := `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4) RETURNING id, user_id, expires_at, created_at`
	err := r.db.QueryRowContext(ctx, query, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return &session, nil
}

func (r *pgRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
