package notification

import (
	"time"

	"github.com/google/uuid"
)

type Batch struct {
	ID         string    `gorm:"primaryKey;type:uuid"`
	Total      int       `gorm:"not null"`
	Pending    int       `gorm:"default:0"`
	Queued     int       `gorm:"default:0"`
	Processing int       `gorm:"default:0"`
	Sent       int       `gorm:"default:0"`
	Failed     int       `gorm:"default:0"`
	Cancelled  int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func NewBatch(total int) *Batch {
	now := time.Now().UTC()
	return &Batch{
		ID:        uuid.NewString(),
		Total:     total,
		Pending:   total,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Apply updates in-memory counters when a notification transitions between statuses.
func (b *Batch) Apply(from, to Status) {
	b.decrement(from)
	b.increment(to)
	b.UpdatedAt = time.Now().UTC()
}

func (b *Batch) increment(s Status) {
	switch s {
	case StatusPending:
		b.Pending++
	case StatusQueued:
		b.Queued++
	case StatusProcessing:
		b.Processing++
	case StatusSent:
		b.Sent++
	case StatusFailed:
		b.Failed++
	case StatusCancelled:
		b.Cancelled++
	}
}

func (b *Batch) decrement(s Status) {
	switch s {
	case StatusPending:
		b.Pending--
	case StatusQueued:
		b.Queued--
	case StatusProcessing:
		b.Processing--
	case StatusSent:
		b.Sent--
	case StatusFailed:
		b.Failed--
	case StatusCancelled:
		b.Cancelled--
	}
}
