package handlers

import (
	"net/http"
	"os"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const sessionCookieName = "session_id"
const sessionMaxAge = 604800 // 7 дней

func Logout(service auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(sessionCookieName)
		if err == nil {
			if sessionID, err := uuid.Parse(cookie); err == nil {
				_ = service.Logout(c.Request.Context(), sessionID)
			}
		}

		clearSessionCookie(c)
		c.Status(http.StatusNoContent)
	}
}

// --- Утилиты для работы с Cookie (используются во всех хендлерах) ---
func setSessionCookie(c *gin.Context, sessionID string) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, sessionID, sessionMaxAge, "/", "", isSecure, true)
}

func clearSessionCookie(c *gin.Context) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", isSecure, true)
}
