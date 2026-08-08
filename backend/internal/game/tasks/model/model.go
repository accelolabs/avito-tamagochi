package model

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	View                  Type = "view"
	Favorite              Type = "favorite"
	CreateListing         Type = "create_listing"
	CreateListingCategory Type = "create_listing_in_category"
)

type Progress struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	LocalDate     time.Time
	TaskType      Type
	Progress      int
	RequiredCount int
	Completed     bool
}
