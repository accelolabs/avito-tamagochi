package main

import (
	"database/sql"
	"github.com/accelolabs/avito-tamagochi/backend/internal/realtime"
	"log"
	"net/http"
	"os"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/handlers"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Читаем конфигурацию
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// 2. Инициализируем подключение к БД
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not parse database URL: %v", err)
	}
	defer db.Close()

	// Обязательно проверяем физическое соединение
	if err := db.Ping(); err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	authRepo := auth.NewPgRepository(db)
	authService := auth.NewService(db, authRepo)

	wsHub := realtime.NewHub()
	go wsHub.Run()

	// 4. Настраиваем роутер
	router := gin.Default()

	wsGroup := router.Group("/")
	wsGroup.Use(middleware.RequireAuth(authService))
	{
		wsGroup.GET("/ws", realtime.ServeWS(wsHub))
	}

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", handlers.Register(authService))
			authGroup.POST("/login", handlers.Login(authService))
			authGroup.POST("/logout", handlers.Logout(authService))
		}

		protected := v1.Group("/")
		protected.Use(middleware.RequireAuth(authService))
		{
			protected.GET("/me", func(c *gin.Context) {
				userID := c.MustGet("userID")
				c.JSON(http.StatusOK, gin.H{
					"message": "You are authenticated!",
					"user_id": userID,
				})
			})

			// Места для будущих эндпоинтов от других бэкендеров:
			// protected.GET("/pet", petHandler.GetPet)
			// protected.POST("/pet/actions", petHandler.PerformAction)
			// protected.GET("/tasks", taskHandler.GetTasks)
			// protected.POST("/demo/activities", demoHandler.SimulateActivity)
		}
	}

	log.Printf("server started on http://0.0.0.0:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
