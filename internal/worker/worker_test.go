package worker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	domain "study-case/internal/domain/notification"
	"study-case/internal/worker"
)

func newWorker(repo *mockRepository, provider *mockProvider, q *mockQueue) *worker.Worker {
	return worker.NewWorker(worker.Config{
		Concurrency:  1,
		MaxRetries:   3,
		PollInterval: 0,
	}, q, repo, provider)
}

func TestProcess_Success(t *testing.T) {
	n := domain.NewNotification("+901111111", domain.ChannelSMS, "test", domain.PriorityNormal)

	repo := &mockRepository{
		UpdateStatusFn: func(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error {
			return nil
		},
	}
	provider := &mockProvider{
		SendFn: func(ctx context.Context, n *domain.Notification) (string, error) {
			return "test-123", nil
		},
	}
	ctx := context.Background()

	w := newWorker(repo, provider, &mockQueue{})
	w.Process(ctx, n)

	assert.Equal(t, domain.StatusSent, n.Status)
	assert.Equal(t, "test-123", *n.ProviderID)
}
