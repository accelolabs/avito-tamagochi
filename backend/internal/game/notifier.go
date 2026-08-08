package game

import "github.com/google/uuid"

type Notifier interface {
	NotifyUser(userID uuid.UUID, event string)
}
