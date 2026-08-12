package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/uuid"
)

const seedUserCount = 20

func Seed(ctx context.Context, db *sql.DB) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC()
	created := 0

	for i := 1; i <= seedUserCount; i++ {
		email := fmt.Sprintf("seed-%d@example.com", i)
		displayName := fmt.Sprintf("Seed_%d", i)
		userID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(email))
		petID := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(fmt.Sprintf("pet:%s", email)))

		var exists bool
		err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email,
		).Scan(&exists)
		if err != nil {
			return 0, err
		}
		if exists {
			continue
		}

		xp := rand.IntN(491) + 10
		chargedAgo := time.Duration(rand.IntN(24)) * time.Hour
		lastChargedAt := now.Add(-chargedAgo)
		streak := rand.IntN(7) + 1
		streakDate := now.Truncate(24 * time.Hour)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (id, email, display_name, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, email, displayName, "$argon2id$seed$notavalidhash", now.Add(-7*24*time.Hour))
		if err != nil {
			return 0, fmt.Errorf("insert user %d: %w", i, err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO pets (id, user_id, xp, last_charged_at, charge_streak, last_streak_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		`, petID, userID, xp, lastChargedAt, streak, streakDate, now.Add(-7*24*time.Hour))
		if err != nil {
			return 0, fmt.Errorf("insert pet %d: %w", i, err)
		}

		eventCount := rand.IntN(5) + 2
		remainingXP := xp
		for j := 0; j < eventCount && remainingXP > 0; j++ {
			eventXP := remainingXP
			if j < eventCount-1 {
				eventXP = rand.IntN(remainingXP) + 1
			}
			remainingXP -= eventXP

			daysAgo := rand.IntN(14)
			eventTime := now.Add(-time.Duration(daysAgo) * 24 * time.Hour)
			eventDate := eventTime.Truncate(24 * time.Hour)
			sourceKey := fmt.Sprintf("seed:%d:event:%d", i, j)

			_, err = tx.ExecContext(ctx, `
				INSERT INTO xp_events (id, user_id, pet_id, source, source_key, amount, occurred_at, local_date)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, uuid.New(), userID, petID, "seed", sourceKey, eventXP, eventTime, eventDate.Format("2006-01-02"))
			if err != nil {
				return 0, fmt.Errorf("insert xp event %d/%d: %w", i, j, err)
			}
		}

		created++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return created, nil
}
