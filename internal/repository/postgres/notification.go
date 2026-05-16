package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"study-case/internal/domain/notification"
)

type NotificationRepository struct {
	db *gorm.DB
}

func (r *NotificationRepository) UpdateStatusBatch(ctx context.Context, id string, status notification.Status) error {
	//TODO implement me
	panic("implement me")
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*notification.Notification, batch *notification.Batch) error {
	panic("not implemented")
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	var n notification.Notification
	err := r.db.WithContext(ctx).First(&n, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, notification.ErrNotFound
	}
	return &n, err
}

func (r *NotificationRepository) List(ctx context.Context, filter notification.Filter) ([]*notification.Notification, int, error) {
	var results []*notification.Notification
	var total int64

	q := r.db.WithContext(ctx).Model(&notification.Notification{})

	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.Channel != nil {
		q = q.Where("channel = ?", *filter.Channel)
	}
	if filter.BatchID != nil {
		q = q.Where("batch_id = ?", *filter.BatchID)
	}
	if filter.StartDate != nil {
		q = q.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		q = q.Where("created_at <= ?", *filter.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := q.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&results).Error
	return results, int(total), err
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status notification.Status, opts ...notification.UpdateOption) error {
	params := &notification.UpdateParams{}
	for _, o := range opts {
		o(params)
	}

	updates := map[string]any{"status": status}
	if params.ProviderID != nil {
		updates["provider_id"] = params.ProviderID
		updates["sent_at"] = time.Now().UTC()
	}
	if params.ErrorMessage != nil {
		updates["error_message"] = params.ErrorMessage
		updates["failed_at"] = time.Now().UTC()
		updates["retry_count"] = gorm.Expr("retry_count + 1")
	}

	return r.db.WithContext(ctx).
		Model(&notification.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *NotificationRepository) Cancel(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&notification.Notification{}).
		Where("id = ? AND status IN ?", id, []notification.Status{notification.StatusPending, notification.StatusQueued}).
		Update("status", notification.StatusCancelled)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var n notification.Notification
		if err := r.db.WithContext(ctx).First(&n, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return notification.ErrNotFound
		}
		return notification.ErrCannotCancel
	}
	return nil
}

func (r *NotificationRepository) GetBatch(ctx context.Context, batchID string) (*notification.Batch, error) {
	panic("not implemented")
}

func (r *NotificationRepository) UpdateBatchCounters(ctx context.Context, batchID string, from, to notification.Status) error {
	panic("not implemented")
}

func (r *NotificationRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&notification.Notification{}).
		Where("idempotency_key = ?", key).
		Count(&count).Error
	return count > 0, err
}
