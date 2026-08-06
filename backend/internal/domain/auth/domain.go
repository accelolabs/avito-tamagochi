package auth

import "time"

// --- REQUEST/RESPONSE DTOs ---

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required"`
	PetName     string `json:"petName" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	User User `json:"user"`
	Pet  Pet  `json:"pet"`
}

// --- DOMAIN MODELS ---

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"` // Omit from JSON responses
	CreatedAt    time.Time `json:"createdAt"`
}

type Pet struct {
	ID                string    `json:"id"`
	OwnerID           string    `json:"ownerId"`
	Name              string    `json:"name"`
	Level             int       `json:"level"`
	XP                int       `json:"xp"`
	XPToNextLevel     int       `json:"xpToNextLevel"`
	BatteryLevel      int       `json:"batteryLevel"`
	Status            string    `json:"status"`
	IsActionAvailable bool      `json:"isActionAvailable"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}
