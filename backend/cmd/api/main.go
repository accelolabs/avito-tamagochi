package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/handler"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	// Setup dependencies
	authRepo := repository.NewPgRepository(db)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	// Setup Gin router
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
		}
	}

	log.Printf("server started on http://0.0.0.0:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
