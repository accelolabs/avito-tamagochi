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

const sessionCookieName = "session_id"

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	resp, session, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	h.setSessionCookie(c, session)
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	session, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	h.setSessionCookie(c, session)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err != nil {
		// If the cookie is not there, the user is already effectively logged out.
		c.Status(http.StatusOK)
		return
	}

	if err := h.service.Logout(c.Request.Context(), sessionID); err != nil {
		// Log the error but don't fail the request, as the goal is to ensure the client is logged out.
		// In a real app, you might want to log this for monitoring.
	}

	h.clearSessionCookie(c)
	c.Status(http.StatusOK)
}

// --- Helper Functions ---

func (h *AuthHandler) setSessionCookie(c *gin.Context, session *auth.Session) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	maxAge := int(time.Until(session.ExpiresAt).Seconds())
	c.SetCookie(sessionCookieName, session.ID, maxAge, "/", "", isSecure, true)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie(sessionCookieName, "", -1, "/", "", isSecure, true)
}
