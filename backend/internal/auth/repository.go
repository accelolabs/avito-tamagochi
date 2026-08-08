package auth

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	return r.findUser(ctx, r.db.QueryRowContext(ctx, `SELECT id, email, display_name, password_hash, created_at FROM users WHERE email = $1`, email))
}

func (r *PostgreSQLRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return r.findUser(ctx, r.db.QueryRowContext(ctx, `SELECT id, email, display_name, password_hash, created_at FROM users WHERE id = $1`, id))
}

func (r *PostgreSQLRepository) findUser(ctx context.Context, row *sql.Row) (*User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgreSQLRepository) CreateUser(ctx context.Context, tx *sql.Tx, user User) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO users (id, email, display_name, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	return err
}

func (r *PostgreSQLRepository) CreateSession(ctx context.Context, tx *sql.Tx, session Session) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4)`, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *PostgreSQLRepository) FindSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	var session Session
	err := r.db.QueryRowContext(ctx, `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1`, id).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *PostgreSQLRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
