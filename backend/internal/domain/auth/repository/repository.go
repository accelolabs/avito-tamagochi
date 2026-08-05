package repository

import (
	"context"
	"database/sql"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
)

type AuthRepository interface {
	CreateUserAndPet(ctx context.Context, user auth.User, pet auth.Pet, session auth.Session) (*auth.User, *auth.Pet, *auth.Session, error)
}

type pgRepository struct {
	db *sql.DB
}

func NewPgRepository(db *sql.DB) AuthRepository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateUserAndPet(ctx context.Context, user auth.User, pet auth.Pet, session auth.Session) (*auth.User, *auth.Pet, *auth.Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback()

	// Create user
	userQuery := `INSERT INTO users (id, email, display_name, password_hash, created_at) 
				   VALUES ($1, $2, $3, $4, $5) RETURNING id, email, display_name, created_at`
	err = tx.QueryRowContext(ctx, userQuery, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create pet
	pet.OwnerID = user.ID
	petQuery := `INSERT INTO pets (id, owner_id, name, level, xp, xp_to_next_level, battery_level, status, is_action_available, updated_at) 
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, owner_id, name, level, xp, xp_to_next_level, battery_level, status, is_action_available, updated_at`
	err = tx.QueryRowContext(ctx, petQuery, pet.ID, pet.OwnerID, pet.Name, pet.Level, pet.XP, pet.XPToNextLevel, pet.BatteryLevel, pet.Status, pet.IsActionAvailable, pet.UpdatedAt).
		Scan(&pet.ID, &pet.OwnerID, &pet.Name, &pet.Level, &pet.XP, &pet.XPToNextLevel, &pet.BatteryLevel, &pet.Status, &pet.IsActionAvailable, &pet.UpdatedAt)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create session
	session.UserID = user.ID
	sessionQuery := `INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3) RETURNING id, user_id, expires_at`
	err = tx.QueryRowContext(ctx, sessionQuery, session.ID, session.UserID, session.ExpiresAt).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, nil, err
	}

	return &user, &pet, &session, nil
}
