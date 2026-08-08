package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/migrations"
	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	directory := os.Getenv("MIGRATIONS_DIR")
	if directory == "" {
		directory = "./migrations"
	}

	if err := migrations.Apply(ctx, db, directory); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	log.Println("database migrations applied")
}
