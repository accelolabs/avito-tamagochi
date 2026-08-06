package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type demoUser struct {
	userID string
	email  string
	name   string
	petID  string
	level  int
	xp     int
}

var demoUsers = []demoUser{
	{userID: "00000000-0000-0000-0000-000000000001", email: "demo.one@example.com", name: "demo_one", petID: "10000000-0000-0000-0000-000000000001", level: 7, xp: 900},
	{userID: "00000000-0000-0000-0000-000000000002", email: "demo.two@example.com", name: "demo_two", petID: "10000000-0000-0000-0000-000000000002", level: 5, xp: 520},
	{userID: "00000000-0000-0000-0000-000000000003", email: "demo.three@example.com", name: "demo_three", petID: "10000000-0000-0000-0000-000000000003", level: 3, xp: 240},
	{userID: "00000000-0000-0000-0000-000000000004", email: "demo.four@example.com", name: "demo_four", petID: "10000000-0000-0000-0000-000000000004", level: 2, xp: 120},
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	switch command {
	case "up":
		err = seed(ctx, db)
	case "cleanup":
		err = cleanup(ctx, db)
	default:
		err = fmt.Errorf("unknown seed command %q; expected up or cleanup", command)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("demo seed command %s completed", command)
}

func seed(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	for _, user := range demoUsers {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, email, display_name, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, display_name = EXCLUDED.display_name`,
			user.userID, user.email, user.name, "demo-password-hash", now); err != nil {
			return fmt.Errorf("seed user %s: %w", user.email, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO pets (id, owner_id, name, level, total_xp, next_level_xp, stage, battery_level, status, is_action_available, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO UPDATE SET level = EXCLUDED.level, total_xp = EXCLUDED.total_xp, updated_at = EXCLUDED.updated_at`,
			user.petID, user.userID, "Кита (demo)", user.level, user.xp, user.xp+100, stageForLevel(user.level), 100, "happy", true, now); err != nil {
			return fmt.Errorf("seed pet for %s: %w", user.email, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}
	return nil
}

func cleanup(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin cleanup transaction: %w", err)
	}
	defer tx.Rollback()

	for _, user := range demoUsers {
		if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.userID); err != nil {
			return fmt.Errorf("remove demo user %s: %w", user.email, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cleanup transaction: %w", err)
	}
	return nil
}

func stageForLevel(level int) string {
	switch {
	case level >= 7:
		return "adult"
	case level >= 4:
		return "teen"
	case level >= 2:
		return "child"
	default:
		return "egg"
	}
}
