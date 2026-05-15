package service

import (
	"context"

	domain "study-case/internal/domain/notification"
)

type notificationService struct {
	repo domain.Repository
}

func New(repo domain.Repository) Service {
	return &notificationService{repo: repo}
}

func (s *notificationService) Create(ctx context.Context, req CreateRequest) (*domain.Notification, error) {
	notification := domain.New(req.Recipient, req.Channel, req.Content, req.Priority)
	err := s.repo.Create(ctx, notification)

	if err != nil {
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
	panic("not implemented")
}

func (s *notificationService) Cancel(ctx context.Context, id string) error {
	panic("not implemented")
}

func (s *notificationService) GetBatch(ctx context.Context, batchID string) (*domain.Batch, error) {
	panic("not implemented")
}
