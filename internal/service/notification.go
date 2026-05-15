package service

import (
	"context"
	"study-case/internal/queue"

	domain "study-case/internal/domain/notification"
)

type notificationService struct {
	repo  domain.Repository
	queue queue.Queue
}

func NewNotificationService(repo domain.Repository, queue queue.Queue) Service {
	return &notificationService{repo: repo, queue: queue}
}

func (s *notificationService) Create(ctx context.Context, req CreateRequest) (*domain.Notification, error) {
	notification := domain.New(req.Recipient, req.Channel, req.Content, req.Priority)
	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	notification.MarkQueued()
	if err := s.repo.UpdateStatus(ctx, notification.ID, domain.StatusQueued); err != nil {
		return nil, err
	}

	if err := s.queue.Enqueue(ctx, notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *notificationService) CreateBatch(ctx context.Context, req CreateBatchRequest) (*BatchResult, error) {
	panic("not implemented")
}

func (s *notificationService) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *notificationService) List(ctx context.Context, filter domain.Filter) ([]*domain.Notification, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *notificationService) Cancel(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}

func (s *notificationService) GetBatch(ctx context.Context, batchID string) (*domain.Batch, error) {
	return s.repo.GetBatch(ctx, batchID)
}
