package queue

import (
	"context"
	"time"

	domain "study-case/internal/domain/notification"
)

type Queue interface {
	Enqueue(ctx context.Context, n *domain.Notification) error
	Dequeue(ctx context.Context, channel domain.Channel) (*domain.Notification, error)
	EnqueueDelayed(ctx context.Context, n *domain.Notification, delay time.Duration) error
	MoveToDead(ctx context.Context, n *domain.Notification) error
}
