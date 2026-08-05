package handler

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, session, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		// Basic error handling
		if errors.Is(err, service.ErrInvalidEmail) || errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Determine if the environment is production for the 'secure' cookie flag
	isSecure := os.Getenv("GIN_MODE") == "release"
	maxAge := int(time.Until(session.ExpiresAt).Seconds())

	c.SetCookie("session_id", session.ID, maxAge, "/", "", isSecure, true)

	c.JSON(http.StatusCreated, resp)
}
