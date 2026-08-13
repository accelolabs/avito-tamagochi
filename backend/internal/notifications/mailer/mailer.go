package mailer

import (
	"context"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
)

type Mailer interface {
	Send(context.Context, notificationmodel.Message) error
}
