package notification

import "errors"

var (
	ErrNotFound                = errors.New("notification not found")
	ErrBatchNotFound           = errors.New("batch not found")
	ErrCannotCancel            = errors.New("notification cannot be cancelled in its current status")
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
	ErrInvalidChannel          = errors.New("invalid channel")
	ErrInvalidPriority         = errors.New("invalid priority")
)
