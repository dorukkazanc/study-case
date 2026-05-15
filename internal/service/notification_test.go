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
	t.Skip("not implemented")
	repo := &mockRepository{}
	s := newService(repo)
	ctx := context.Background()

	// TODO: repo.GetByIDFn → ErrNotFound dönsün
	_, err := s.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// --- List ---

func TestList_Success(t *testing.T) {
	t.Skip("not implemented")

	// TODO: repo.ListFn birkaç notification dönsün
	// List çağır → aynı sonuçlar gelsin
	require.Fail(t, "not implemented")
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
