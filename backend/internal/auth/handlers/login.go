package handlers

import (
	"errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(service auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "invalid input"})
			return
		}

		user, session, err := service.Login(c.Request.Context(), auth.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		})

		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				c.JSON(http.StatusUnauthorized, gin.H{"code": "invalid_credentials", "message": "invalid email or password"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to login"})
			return
		}

		setSessionCookie(c, session.ID.String())
		c.JSON(http.StatusOK, user)
	}
}
