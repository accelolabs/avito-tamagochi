package repository

import (
	"context"
	"database/sql"
	"fmt"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
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

func (r *PostgreSQLRepository) ListParticipants(ctx context.Context) ([]notificationmodel.Participant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.email, p.energy_percent, p.energy_updated_at
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
		if err := rows.Scan(&participant.UserID, &participant.Email, &participant.EnergyPercent, &participant.EnergyUpdatedAt); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}
	return participants, rows.Err()
}

func (r *PostgreSQLRepository) WithParticipantLock(
	ctx context.Context,
	participant notificationmodel.Participant,
	handle notificationmodel.ParticipantHandler,
) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	participant, err = participantForUpdate(ctx, tx, participant)
	if err != nil {
		return false, err
	}
	delivered, err := deliveredThresholds(ctx, tx, participant.UserID)
	if err != nil {
		return false, err
	}
	threshold, err := handle(ctx, participant, delivered)
	if err != nil {
		return false, err
	}
	if threshold != nil {
		if err := recordDelivery(ctx, tx, participant.UserID, *threshold); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return threshold != nil, nil
}

func participantForUpdate(ctx context.Context, tx *sql.Tx, participant notificationmodel.Participant) (notificationmodel.Participant, error) {
	err := tx.QueryRowContext(ctx, `
		SELECT u.email, p.energy_percent, p.energy_updated_at
		FROM users u
		JOIN pets p ON p.user_id = u.id
		WHERE u.id = $1
		FOR UPDATE OF p
	`, participant.UserID).Scan(&participant.Email, &participant.EnergyPercent, &participant.EnergyUpdatedAt)
	return participant, err
}

func deliveredThresholds(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (notificationmodel.DeliveredThresholds, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT threshold FROM energy_notification_deliveries WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	delivered := make(notificationmodel.DeliveredThresholds)
	for rows.Next() {
		var threshold int
		if err := rows.Scan(&threshold); err != nil {
			return nil, err
		}
		delivered[threshold] = true
	}
	return delivered, rows.Err()
}

func recordDelivery(ctx context.Context, tx *sql.Tx, userID uuid.UUID, threshold int) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO energy_notification_deliveries (user_id, threshold, sent_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
	`, userID, threshold)
	if err != nil {
		return fmt.Errorf("record notification delivery: %w", err)
	}
	return nil
}
