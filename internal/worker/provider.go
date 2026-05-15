package worker

import (
	"context"
	"study-case/internal/domain/notification"
)

type Provider interface {
	Send(ctx context.Context, notification *notification.Notification) (providerId string, err error)
}
