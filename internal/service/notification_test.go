package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "study-case/internal/domain/notification"
	"study-case/internal/service"
)

func newService(repo *mockRepository, queue *mockQueue) service.Service {
	return service.NewNotificationService(repo, queue, zerolog.Nop())
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepository{
		CreateFn: func(ctx context.Context, n *domain.Notification) error {
			return nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status domain.Status, opts ...domain.UpdateOption) error {
			return nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	result, err := s.Create(ctx, service.CreateRequest{
		Recipient: "+9011111111",
		Channel:   domain.ChannelSMS,
		Content:   "test",
		Priority:  domain.PriorityNormal,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.StatusQueued, result.Status)
	assert.Equal(t, "+9011111111", result.Recipient)
}

func TestGetByID_Success(t *testing.T) {
	expectedId := uuid.NewString()
	repo := &mockRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Notification, error) {
			return &domain.Notification{
				ID: expectedId,
			}, nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	result, err := s.GetByID(ctx, expectedId)

	require.NoError(t, err)
	assert.Equal(t, result.ID, expectedId)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Notification, error) {
			return nil, domain.ErrNotFound
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	_, err := s.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestList_Success(t *testing.T) {
	n1 := domain.NewNotification("1", domain.ChannelSMS, "mock1", domain.PriorityNormal)
	n2 := domain.NewNotification("2", domain.ChannelSMS, "mock2", domain.PriorityNormal)

	repo := &mockRepository{
		ListFn: func(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error) {
			return []*domain.Notification{n1, n2}, 2, nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	results, total, err := s.List(ctx, domain.Filter{
		Page:     1,
		PageSize: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)
	assert.Equal(t, n1, results[0])
	assert.Equal(t, n2, results[1])
}

func TestCancel_Success(t *testing.T) {
	repo := &mockRepository{
		CancelFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	err := s.Cancel(ctx, uuid.NewString())
	assert.NoError(t, err)
}

func TestCancel_NotFound(t *testing.T) {
	repo := &mockRepository{
		CancelFn: func(ctx context.Context, id string) error {
			return domain.ErrNotFound
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	err := s.Cancel(ctx, uuid.NewString())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCancel_CannotCancel(t *testing.T) {
	repo := &mockRepository{
		CancelFn: func(ctx context.Context, id string) error {
			return domain.ErrCannotCancel
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	err := s.Cancel(ctx, uuid.NewString())
	assert.ErrorIs(t, err, domain.ErrCannotCancel)
}

func TestGetBatch_Success(t *testing.T) {
	batchID := uuid.NewString()
	repo := &mockRepository{
		GetBatchFn: func(ctx context.Context, batchID string) (*domain.Batch, error) {
			return &domain.Batch{
				ID:        batchID,
				Failed:    0,
				Cancelled: 0,
				Sent:      10,
				Total:     10,
				Queued:    0,
			}, nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	result, err := s.GetBatch(ctx, batchID)

	require.NoError(t, err)
	assert.Equal(t, 10, result.Total)
	assert.Equal(t, 0, result.Queued)
	assert.Equal(t, 10, result.Sent)
	assert.Equal(t, batchID, result.ID)
}

func TestGetBatch_NotFound(t *testing.T) {
	repo := &mockRepository{
		GetBatchFn: func(ctx context.Context, batchID string) (*domain.Batch, error) {
			return nil, domain.ErrBatchNotFound
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	_, err := s.GetBatch(ctx, uuid.NewString())
	assert.ErrorIs(t, err, domain.ErrBatchNotFound)
}

func TestCreate_DuplicateIdempotencyKey(t *testing.T) {
	repo := &mockRepository{
		ExistsByIdempotencyKeyFn: func(ctx context.Context, key string) (bool, error) {
			return true, nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()
	_, err := s.Create(ctx, service.CreateRequest{
		Recipient:      "+9011111111",
		Channel:        domain.ChannelSMS,
		Content:        "test",
		Priority:       domain.PriorityNormal,
		IdempotencyKey: new("testkey"),
	})
	require.ErrorIs(t, err, domain.ErrDuplicateIdempotencyKey)
}

func TestCreateBatch_Success(t *testing.T) {
	repo := &mockRepository{
		CreateBatchFn: func(ctx context.Context, notifications []*domain.Notification, batch *domain.Batch) error {
			return nil
		},
		UpdateStatusBatchFn: func(ctx context.Context, batchID string, status domain.Status) error {
			return nil
		},
	}
	queue := &mockQueue{}

	s := newService(repo, queue)
	ctx := context.Background()

	result, err := s.CreateBatch(ctx, service.CreateBatchRequest{
		Notifications: []service.CreateRequest{
			{Recipient: "1", Channel: domain.ChannelSMS, Content: "mock1", Priority: domain.PriorityNormal},
			{Recipient: "2", Channel: domain.ChannelSMS, Content: "mock2", Priority: domain.PriorityNormal},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Count)
	assert.NotEmpty(t, result.BatchID)
}

func TestCreate_Scheduled_Success(t *testing.T) {
	enqueueDelayed := false
	repo := &mockRepository{}
	queue := &mockQueue{
		EnqueueFn: func(ctx context.Context, n *domain.Notification) error {
			t.Fatal("Enqueue should not be called for scheduled notification")
			return nil
		},
		EnqueueDelayedFn: func(ctx context.Context, n *domain.Notification, delay time.Duration) error {
			enqueueDelayed = true
			return nil
		},
	}

	s := newService(repo, queue)
	ctx := context.Background()

	scheduledAt := time.Now().Add(20 * time.Second)
	result, err := s.Create(ctx, service.CreateRequest{
		Recipient:   "+9011111111",
		Channel:     domain.ChannelSMS,
		Content:     "test",
		Priority:    domain.PriorityNormal,
		ScheduledAt: &scheduledAt,
	})

	require.NoError(t, err)
	assert.True(t, enqueueDelayed)
	assert.NotNil(t, result.ScheduledAt)
}
