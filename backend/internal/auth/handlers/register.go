package handlers

import (
	"errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var displayNamePattern = regexp.MustCompile(`^[A-Za-z_-]+$`)

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email,max=254"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	DisplayName string `json:"displayName" binding:"required,min=2,max=32"`
}

func Register(service auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": err.Error()})
			return
		}

		if !displayNamePattern.MatchString(req.DisplayName) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "invalid characters in displayName"})
			return
		}

		input := auth.RegisterInput{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		}

		user, session, err := service.Register(c.Request.Context(), input)
		if err != nil {
			if errors.Is(err, auth.ErrEmailAlreadyExists) {
				c.JSON(http.StatusConflict, gin.H{"code": "email_already_exists", "message": "email already registered"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to register"})
			return
		}

		setSessionCookie(c, session.ID.String())
		c.JSON(http.StatusCreated, user)
	}
}
