package service_test

import (
	"context"
	"study-case/internal/domain/notification"
	"time"

	domain "study-case/internal/domain/notification"
)

type mockRepository struct {
	CreateFn                 func(ctx context.Context, n *domain.Notification) error
	CreateBatchFn            func(ctx context.Context, notifications []*domain.Notification, batch *domain.Batch) error
	GetByIDFn                func(ctx context.Context, id string) (*domain.Notification, error)
	ListFn                   func(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error)
	UpdateStatusFn           func(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error
	CancelFn                 func(ctx context.Context, id string) error
	GetBatchFn               func(ctx context.Context, batchID string) (*domain.Batch, error)
	UpdateBatchCountersFn    func(ctx context.Context, batchID string, from, to domain.Status) error
	ExistsByIdempotencyKeyFn func(ctx context.Context, key string) (bool, error)
}

func (m *mockRepository) UpdateStatusBatch(ctx context.Context, id string, status notification.Status) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockRepository) Create(ctx context.Context, n *domain.Notification) error {
	return m.CreateFn(ctx, n)
}

func (m *mockRepository) CreateBatch(ctx context.Context, notifications []*domain.Notification, batch *domain.Batch) error {
	return m.CreateBatchFn(ctx, notifications, batch)
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	return m.GetByIDFn(ctx, id)
}

func (m *mockRepository) List(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error) {
	return m.ListFn(ctx, filter)
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error {
	return m.UpdateStatusFn(ctx, id, status, opts...)
}

func (m *mockRepository) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *mockRepository) GetBatch(ctx context.Context, batchID string) (*domain.Batch, error) {
	return m.GetBatchFn(ctx, batchID)
}

func (m *mockRepository) UpdateBatchCounters(ctx context.Context, batchID string, from, to domain.Status) error {
	return m.UpdateBatchCountersFn(ctx, batchID, from, to)
}

func (m *mockRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return m.ExistsByIdempotencyKeyFn(ctx, key)
}

type mockQueue struct {
}

func (m *mockQueue) Enqueue(ctx context.Context, n *domain.Notification) error { return nil }
func (m *mockQueue) Dequeue(ctx context.Context, ch domain.Channel) (*domain.Notification, error) {
	return nil, nil
}
func (m *mockQueue) EnqueueDelayed(ctx context.Context, n *domain.Notification, delay time.Duration) error {
	return nil
}
func (m *mockQueue) MoveToDead(ctx context.Context, n *domain.Notification) error { return nil }
