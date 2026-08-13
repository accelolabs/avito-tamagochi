package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mailer "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/mailer"
	repository "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/repository"
	service "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/service"
	_ "github.com/lib/pq"
)

const smtpTimeout = 5 * time.Second

func main() { os.Exit(run()) }

func run() int {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Print("DATABASE_URL environment variable is not set")
		return 1
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Printf("open database: %v", err)
		return 1
	}
	defer func() { _ = db.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		log.Printf("ping database: %v", err)
		return 1
	}

	dispatcher := service.New(
		repository.New(db),
		mailer.NewSMTP(environmentOrDefault("SMTP_ADDRESS", "localhost:1025"), environmentOrDefault("MAIL_FROM", "no-reply@tamagochi.local"), smtpTimeout),
	)
	result, err := dispatcher.DispatchAll(ctx)
	log.Printf("energy notifications: participants=%d sent=%d skipped=%d failed=%d", result.Participants, result.Sent, result.Skipped, result.Failed)
	if err != nil {
		log.Printf("dispatch energy notifications: %v", err)
		return 1
	}
	return 0
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
