package worker

import (
	"context"
	domain "study-case/internal/domain/notification"
)

type Repository interface {
	UpdateStatus(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error
	UpdateBatchCounters(ctx context.Context, batchID string, from, to domain.Status) error
}
