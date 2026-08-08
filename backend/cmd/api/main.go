package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	authhandlers "github.com/accelolabs/avito-tamagochi/backend/internal/auth/handlers"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/service"
	"github.com/accelolabs/avito-tamagochi/backend/internal/realtime"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	authService := service.NewService(db, repository.NewRepository(db))
	authHandler := authhandlers.New(authService)
	hub := realtime.NewHub()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	authHandler.SetRoutes(mux)
	mux.Handle("GET /api/v1/auth/me", middleware.RequireSession(authService, http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /ws", middleware.RequireSession(authService, realtime.ServeWS(hub)))

	log.Printf("server started on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
