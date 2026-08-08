package auth

import (
	"context"
	"database/sql"
	
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type pgRepository struct {
	db *sql.DB
}

func NewPgRepository(db *sql.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateUser(ctx context.Context, tx *sql.Tx, user User) error {
	query := `
		INSERT INTO users (id, email, display_name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(ctx, query, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	if err != nil {
		// Перехват ошибки уникальности из PostgreSQL (код 23505)
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *pgRepository) CreateSession(ctx context.Context, tx *sql.Tx, session Session) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *pgRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, display_name, password_hash, created_at FROM users WHERE email = $1`
	var user User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *pgRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, email, display_name, password_hash, created_at FROM users WHERE id = $1`
	var user User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *pgRepository) FindSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	query := `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1`
	var s Session
	err := r.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
