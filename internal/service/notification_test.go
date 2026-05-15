package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "study-case/internal/domain/notification"
	"study-case/internal/service"
)

func newService(repo *mockRepository) service.Service {
	return service.New(repo)
}

func TestCreate_Success(t *testing.T) {
	repo := &mockRepository{
		CreateFn: func(ctx context.Context, n *domain.Notification) error {
			return nil
		},
	}
	s := newService(repo)
	ctx := context.Background()

	result, err := s.Create(ctx, service.CreateRequest{
		Recipient: "+9011111111",
		Channel:   domain.ChannelSMS,
		Content:   "test",
		Priority:  domain.PriorityNormal,
	})

	require.NoError(t, err)
	assert.Equal(t, domain.StatusPending, result.Status)
	assert.Equal(t, "+9011111111", result.Recipient)
}

func TestGetByID_Success(t *testing.T) {
	expectedId := uuid.New().String()
	repo := &mockRepository{
		GetByIDFn: func(ctx context.Context, id string) (*domain.Notification, error) {
			return &domain.Notification{
				ID: expectedId,
			}, nil
		},
	}
	s := newService(repo)
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
	s := newService(repo)
	ctx := context.Background()

	_, err := s.GetByID(ctx, uuid.New().String())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// --- List ---

func TestList_Success(t *testing.T) {
	n1 := domain.New("1", domain.ChannelSMS, "mock1", domain.PriorityNormal)
	n2 := domain.New("2", domain.ChannelSMS, "mock2", domain.PriorityNormal)

	repo := &mockRepository{
		ListFn: func(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error) {
			return []*domain.Notification{n1, n2}, 2, nil
		},
	}
	s := newService(repo)
	ctx := context.Background()

	results, total, err := s.List(ctx, domain.Filter{
		Page:     1,
		PageSize: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)
}

// --- Cancel ---

func TestCancel_Success(t *testing.T) {
	t.Skip("not implemented")
	repo := &mockRepository{}
	s := newService(repo)
	ctx := context.Background()

	// TODO: repo.CancelFn nil dönsün
	err := s.Cancel(ctx, "some-id")
	assert.NoError(t, err)
}

func TestCancel_NotFound(t *testing.T) {
	t.Skip("not implemented")
	repo := &mockRepository{}
	s := newService(repo)
	ctx := context.Background()

	// TODO: repo.CancelFn → ErrNotFound dönsün
	err := s.Cancel(ctx, "non-existent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestCancel_CannotCancel(t *testing.T) {
	t.Skip("not implemented")
	repo := &mockRepository{}
	s := newService(repo)
	ctx := context.Background()

	// TODO: repo.CancelFn → ErrCannotCancel dönsün
	err := s.Cancel(ctx, "sent-id")
	assert.ErrorIs(t, err, domain.ErrCannotCancel)
}

// --- GetBatch ---

func TestGetBatch_Success(t *testing.T) {
	t.Skip("not implemented")

	// TODO: repo.GetBatchFn bir batch dönsün
	require.Fail(t, "not implemented")
}

func TestGetBatch_NotFound(t *testing.T) {
	t.Skip("not implemented")
	repo := &mockRepository{}
	s := newService(repo)
	ctx := context.Background()

	// TODO: repo.GetBatchFn → ErrBatchNotFound dönsün
	_, err := s.GetBatch(ctx, "non-existent")
	assert.ErrorIs(t, err, domain.ErrBatchNotFound)
}
