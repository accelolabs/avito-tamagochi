package tasks

import (
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskView                  TaskType = "view"
	TaskFavorite              TaskType = "favorite"
	TaskCreateListing         TaskType = "create_listing"
	TaskCreateListingCategory TaskType = "create_listing_in_category"
)

type TaskProgress struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	LocalDate     time.Time
	TaskType      TaskType
	Progress      int
	RequiredCount int
	Completed     bool
}
