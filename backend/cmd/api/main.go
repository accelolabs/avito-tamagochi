package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	authhandlers "github.com/accelolabs/avito-tamagochi/backend/internal/auth/handlers"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	authrepository "github.com/accelolabs/avito-tamagochi/backend/internal/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/service"
	gamehandlers "github.com/accelolabs/avito-tamagochi/backend/internal/game/handlers"
	gameleaderrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/repository"
	gameleaderservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/leaderboard/service"
	gamepetrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/repository"
	gamepetservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	gameprogressionrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/repository"
	gameprogressionservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/progression/service"
	gamerewardrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/repository"
	gamerewardservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/service"
	gamesummaryrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/repository"
	gamesummaryservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/summary/service"
	gametaskrepository "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/repository"
	gametaskservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/tasks/service"
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

	authService := service.NewService(db, authrepository.New(db))
	authHandler := authhandlers.New(authService)
	hub := realtime.NewHub()
	go hub.Run()
	progressionService := gameprogressionservice.New()
	petRepo := gamepetrepository.New(db)
	xpRepo := gameprogressionrepository.New(db)
	rewardRepo := gamerewardrepository.New(db)
	petService := gamepetservice.New(db, petRepo, xpRepo, rewardRepo, nil, progressionService, hub)
	taskService := gametaskservice.New(db, gametaskrepository.New(db), petRepo, xpRepo, rewardRepo, nil, progressionService, hub)
	rewardService := gamerewardservice.New(db, rewardRepo, hub)
	summaryService := gamesummaryservice.New(gamesummaryrepository.New(db), nil)
	leaderboardService := gameleaderservice.New(gameleaderrepository.New(db))
	gameHandler := gamehandlers.New(petService, taskService, rewardService, summaryService, leaderboardService, authService)

	mux := http.NewServeMux()
	authHandler.SetRoutes(mux)
	gameHandler.SetRoutes(mux)
	mux.Handle("GET /api/v1/auth/me", middleware.RequireSession(authService, http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /ws", middleware.RequireSession(authService, realtime.ServeWS(hub)))

	log.Printf("server started on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
