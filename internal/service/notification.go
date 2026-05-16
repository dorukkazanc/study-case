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
	notification := domain.NewNotification(req.Recipient, req.Channel, req.Content, req.Priority)
	if req.IdempotencyKey != nil {
		exists, err := s.repo.ExistsByIdempotencyKey(ctx, *req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrDuplicateIdempotencyKey
		}
	}
	notification.IdempotencyKey = req.IdempotencyKey

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
	list := createNotificationListFromBatchRequest(req)
	newBatch := domain.NewBatch(len(list))
	for _, notification := range list {
		notification.BatchID = &newBatch.ID
	}

	err := s.repo.CreateBatch(ctx, list, newBatch)

	if err != nil {
		return nil, err
	}
	return &BatchResult{
		BatchID: newBatch.ID,
		Count:   len(list),
	}, nil
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

func createNotificationListFromBatchRequest(req CreateBatchRequest) []*domain.Notification {
	notifications := make([]*domain.Notification, 0, len(req.Notifications))
	for _, r := range req.Notifications {
		n := domain.NewNotification(r.Recipient, r.Channel, r.Content, r.Priority)
		n.IdempotencyKey = r.IdempotencyKey
		notifications = append(notifications, n)
	}
	return notifications
}
