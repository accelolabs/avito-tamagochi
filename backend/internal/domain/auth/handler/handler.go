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

const (
	sessionCookieName = "session_id"
	sessionMaxAge     = 604800 // 7 days in seconds
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
		c.JSON(http.StatusBadRequest, auth.ErrorResponse{Code: "validation_error", Message: err.Error()})
		return
	}

	user, session, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, auth.ErrorResponse{Code: "email_already_exists", Message: "Email is already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, auth.ErrorResponse{Code: "internal_error", Message: "Failed to register user"})
		return
	}

	h.setSessionCookie(c, session)
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, auth.ErrorResponse{Code: "validation_error", Message: err.Error()})
		return
	}

	user, session, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, auth.ErrorResponse{Code: "invalid_credentials", Message: "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, auth.ErrorResponse{Code: "internal_error", Message: "Failed to login"})
		return
	}

	h.setSessionCookie(c, session)
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err == nil { // Only attempt to delete from DB if cookie exists
		if err := h.service.Logout(c.Request.Context(), sessionID); err != nil {
			// Log the error but don't fail the request, as the goal is to ensure the client is logged out.
			// In a real app, you might want to log this for monitoring.
			c.Error(err) // Use c.Error to log the error with Gin's logger
		}
	}

	h.clearSessionCookie(c)
	c.Status(http.StatusNoContent) // 204 No Content as per spec
}

// --- Helper Functions ---

func (h *AuthHandler) setSessionCookie(c *gin.Context, session *auth.Session) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	// MaxAge is calculated from now until expiration, ensuring it's always positive.
	// If session.ExpiresAt is in the past, MaxAge will be 0 or negative, effectively expiring the cookie.
	maxAge := int(time.Until(session.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetCookie(sessionCookieName, session.ID, maxAge, "/", "", isSecure, true)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetCookie(sessionCookieName, "", -1, "/", "", isSecure, true) // MaxAge -1 clears the cookie
}
