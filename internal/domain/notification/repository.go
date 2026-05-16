package notification

import "context"

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	CreateBatch(ctx context.Context, notifications []*Notification, batch *Batch) error
	GetByID(ctx context.Context, id string) (*Notification, error)
	List(ctx context.Context, filter Filter) ([]*Notification, int, error)
	UpdateStatus(ctx context.Context, id string, status Status, opts ...UpdateOption) error
	Cancel(ctx context.Context, id string) error
	GetBatch(ctx context.Context, batchID string) (*Batch, error)
	UpdateBatchCounters(ctx context.Context, batchID string, from, to Status) error
	ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
	UpdateStatusBatch(ctx context.Context, id string, status Status) error
}

type UpdateOption func(*UpdateParams)

type UpdateParams struct {
	ProviderID   *string
	ErrorMessage *string
}

func WithProviderID(id string) UpdateOption {
	return func(p *UpdateParams) { p.ProviderID = &id }
}

func WithErrorMessage(msg string) UpdateOption {
	return func(p *UpdateParams) { p.ErrorMessage = &msg }
}
