package auth

import "time"

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	PetName     string `json:"petName"`
}

type AuthResponse struct {
	User User `json:"user"`
	Pet  Pet  `json:"pet"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string
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
