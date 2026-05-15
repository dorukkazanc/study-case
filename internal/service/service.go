package service

import (
	"context"
	"time"

	domain "study-case/internal/domain/notification"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*domain.Notification, error)
	CreateBatch(ctx context.Context, req CreateBatchRequest) (*BatchResult, error)
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	List(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error)
	Cancel(ctx context.Context, id string) error
	GetBatch(ctx context.Context, batchID string) (*domain.Batch, error)
}

type CreateRequest struct {
	Recipient      string
	Channel        domain.Channel
	Content        string
	Priority       domain.Priority
	IdempotencyKey *string
	ScheduledAt    *time.Time
}

type CreateBatchRequest struct {
	Notifications []CreateRequest
}

type BatchResult struct {
	BatchID string
	Count   int
}
