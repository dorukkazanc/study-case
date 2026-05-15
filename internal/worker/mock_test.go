package worker_test

import (
	"context"
	"time"

	domain "study-case/internal/domain/notification"
)

type mockRepository struct {
	UpdateStatusFn        func(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error
	UpdateBatchCountersFn func(ctx context.Context, batchID string, from, to domain.Status) error
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error {
	return m.UpdateStatusFn(ctx, id, status, opts...)
}

func (m *mockRepository) UpdateBatchCounters(ctx context.Context, batchID string, from, to domain.Status) error {
	return m.UpdateBatchCountersFn(ctx, batchID, from, to)
}

type mockProvider struct {
	SendFn func(ctx context.Context, n *domain.Notification) (string, error)
}

func (m *mockProvider) Send(ctx context.Context, n *domain.Notification) (string, error) {
	return m.SendFn(ctx, n)
}

type mockQueue struct {
	EnqueueFn        func(ctx context.Context, n *domain.Notification) error
	DequeueFn        func(ctx context.Context, channel domain.Channel) (*domain.Notification, error)
	EnqueueDelayedFn func(ctx context.Context, n *domain.Notification, delay time.Duration) error
	MoveToDeadFn     func(ctx context.Context, n *domain.Notification) error
}

func (m *mockQueue) Enqueue(ctx context.Context, n *domain.Notification) error {
	if m.EnqueueFn != nil {
		return m.EnqueueFn(ctx, n)
	}
	return nil
}

func (m *mockQueue) Dequeue(ctx context.Context, channel domain.Channel) (*domain.Notification, error) {
	if m.DequeueFn != nil {
		return m.DequeueFn(ctx, channel)
	}
	return nil, nil
}

func (m *mockQueue) EnqueueDelayed(ctx context.Context, n *domain.Notification, delay time.Duration) error {
	if m.EnqueueDelayedFn != nil {
		return m.EnqueueDelayedFn(ctx, n, delay)
	}
	return nil
}

func (m *mockQueue) MoveToDead(ctx context.Context, n *domain.Notification) error {
	if m.MoveToDeadFn != nil {
		return m.MoveToDeadFn(ctx, n)
	}
	return nil
}
