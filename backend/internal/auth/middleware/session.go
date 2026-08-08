package middleware

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireAuth(service auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "missing session cookie",
			})
			return
		}

		sessionID, err := uuid.Parse(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "invalid session format",
			})
			return
		}

		userID, err := service.ValidateSession(c.Request.Context(), sessionID)
		if err != nil {
			if errors.Is(err, auth.ErrSessionNotFound) || errors.Is(err, auth.ErrSessionExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    "unauthorized",
					"message": "session expired or invalid",
				})
				return
			}
			// При ошибке базы данных не раскрываем детали
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "internal_error",
				"message": "failed to validate session",
			})
			return
		}

		// Кладем userID в контекст для следующих хендлеров
		c.Set("userID", userID)
		c.Next()
	}
}
