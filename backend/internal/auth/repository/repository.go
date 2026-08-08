package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
)

type PostgreSQLRepository struct{ db *sql.DB }

type User = auth.User
type Session = auth.Session

var ErrEmailAlreadyExists = auth.ErrEmailAlreadyExists

func NewRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) FindUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT id, email, display_name, password_hash, created_at FROM users WHERE email = $1`, email))
}

func (r *PostgreSQLRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT id, email, display_name, password_hash, created_at FROM users WHERE id = $1`, id))
}

func scanUser(row *sql.Row) (*auth.User, error) {
	var user auth.User
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgreSQLRepository) CreateUser(ctx context.Context, tx *sql.Tx, user auth.User) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO users (id, email, display_name, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	if isUniqueViolation(err) {
		return auth.ErrEmailAlreadyExists
	}
	return err
}

func (r *PostgreSQLRepository) CreateSession(ctx context.Context, tx *sql.Tx, session auth.Session) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4)`, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

const uniqueViolationCode = "23505"

func isUniqueViolation(err error) bool {
	var pgErr *pq.Error
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

func (r *PostgreSQLRepository) FindSession(ctx context.Context, id uuid.UUID) (*auth.Session, error) {
	var session auth.Session
	err := r.db.QueryRowContext(ctx, `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1`, id).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *PostgreSQLRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}
