package service

import (
	"context"
	"study-case/internal/metrics"
	"study-case/internal/queue"
	"time"

	domain "study-case/internal/domain/notification"

	"github.com/rs/zerolog"
)

type notificationService struct {
	repo  domain.Repository
	queue queue.Queue
	log   zerolog.Logger
}

func NewNotificationService(repo domain.Repository, queue queue.Queue, log zerolog.Logger) Service {
	return &notificationService{repo: repo, queue: queue, log: log}
}

func (s *notificationService) Create(ctx context.Context, req CreateRequest) (*domain.Notification, error) {
	if req.IdempotencyKey != nil {
		exists, err := s.repo.ExistsByIdempotencyKey(ctx, *req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrDuplicateIdempotencyKey
		}
	}

	n := domain.NewNotification(req.Recipient, req.Channel, req.Content, req.Priority)
	n.IdempotencyKey = req.IdempotencyKey
	n.ScheduledAt = req.ScheduledAt

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}

	n.MarkQueued()
	if err := s.repo.UpdateStatus(ctx, n.ID, domain.StatusQueued); err != nil {
		return nil, err
	}

	if err := s.enqueue(ctx, n); err != nil {
		return nil, err
	}

	metrics.NotificationsCreated.WithLabelValues(string(n.Channel), string(n.Priority)).Inc()
	s.log.Info().Str("id", n.ID).Str("channel", string(n.Channel)).Msg("notification created")
	return n, nil
}

func (s *notificationService) enqueue(ctx context.Context, n *domain.Notification) error {
	if n.ScheduledAt != nil {
		return s.queue.EnqueueDelayed(ctx, n, time.Until(*n.ScheduledAt))
	}
	return s.queue.Enqueue(ctx, n)
}

func (s *notificationService) CreateBatch(ctx context.Context, req CreateBatchRequest) (*BatchResult, error) {
	list := createNotificationListFromBatchRequest(req)
	newBatch := domain.NewBatch(len(list))
	for _, notification := range list {
		notification.BatchID = &newBatch.ID
	}

	if err := s.repo.CreateBatch(ctx, list, newBatch); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatusBatch(ctx, newBatch.ID, domain.StatusQueued); err != nil {
		return nil, err
	}

	for _, notification := range list {
		notification.MarkQueued()
		_ = s.queue.Enqueue(ctx, notification)
	}

	s.log.Info().Str("batch_id", newBatch.ID).Int("count", len(list)).Msg("batch created")
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
