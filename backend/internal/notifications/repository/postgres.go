package repository

import (
	"context"
	"database/sql"
	"fmt"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
)

const runLockID int64 = 824015012024

type PostgreSQLRepository struct{ db *sql.DB }

func New(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

func (r *PostgreSQLRepository) TryRunLock(ctx context.Context) (func(), bool, error) {
	connection, err := r.db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var acquired bool
	if err := connection.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, runLockID).Scan(&acquired); err != nil {
		_ = connection.Close()
		return nil, false, err
	}
	if !acquired {
		_ = connection.Close()
		return func() {}, false, nil
	}
	release := func() {
		_, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, runLockID)
		_ = connection.Close()
	}
	return release, true, nil
}

func (r *PostgreSQLRepository) ListParticipantIDs(ctx context.Context) ([]notificationmodel.Participant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.email, p.last_charged_at
		FROM users u
		JOIN pets p ON p.user_id = u.id
		ORDER BY u.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var participants []notificationmodel.Participant
	for rows.Next() {
		var participant notificationmodel.Participant
		if err := rows.Scan(&participant.UserID, &participant.Email, &participant.LastChargedAt); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}
	return participants, rows.Err()
}

func (r *PostgreSQLRepository) ProcessParticipant(
	ctx context.Context,
	participant notificationmodel.Participant,
	dispatch func(notificationmodel.Participant, map[int]bool) (*int, error),
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.QueryRowContext(ctx, `
		SELECT u.email, p.last_charged_at
		FROM users u
		JOIN pets p ON p.user_id = u.id
		WHERE u.id = $1
		FOR UPDATE OF p
	`, participant.UserID).Scan(&participant.Email, &participant.LastChargedAt); err != nil {
		return false, err
	}
	delivered, err := deliveredThresholds(ctx, tx, participant)
	if err != nil {
		return false, err
	}
	threshold, err := dispatch(participant, delivered)
	if err != nil {
		return false, err
	}
	if threshold != nil {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO energy_notification_deliveries (user_id, threshold, sent_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
		`, participant.UserID, *threshold); err != nil {
			return false, fmt.Errorf("record notification delivery: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return threshold != nil, nil
}

func deliveredThresholds(ctx context.Context, tx *sql.Tx, participant notificationmodel.Participant) (map[int]bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT threshold FROM energy_notification_deliveries WHERE user_id = $1
	`, participant.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	delivered := make(map[int]bool)
	for rows.Next() {
		var threshold int
		if err := rows.Scan(&threshold); err != nil {
			return nil, err
		}
		delivered[threshold] = true
	}
	return delivered, rows.Err()
}
