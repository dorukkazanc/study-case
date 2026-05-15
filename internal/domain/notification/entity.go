package notification

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
)

func ParseChannel(s string) (Channel, error) {
	switch Channel(s) {
	case ChannelSMS, ChannelEmail, ChannelPush:
		return Channel(s), nil
	default:
		return "", ErrInvalidChannel
	}
}

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

func ParsePriority(s string) (Priority, error) {
	switch Priority(s) {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return Priority(s), nil
	default:
		return "", ErrInvalidPriority
	}
}

type Status string

const (
	StatusPending    Status = "pending"
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusSent       Status = "sent"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Notification struct {
	ID             string     `gorm:"primaryKey;type:uuid"`
	BatchID        *string    `gorm:"type:uuid;index"`
	Recipient      string     `gorm:"not null"`
	Channel        Channel    `gorm:"type:varchar(20);not null"`
	Content        string     `gorm:"type:text;not null"`
	Priority       Priority   `gorm:"type:varchar(20);not null;default:normal"`
	Status         Status     `gorm:"type:varchar(20);not null;default:pending;index"`
	IdempotencyKey *string    `gorm:"type:varchar(255);uniqueIndex"`
	ScheduledAt    *time.Time `gorm:"index"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
	SentAt         *time.Time
	FailedAt       *time.Time
	RetryCount     int     `gorm:"default:0"`
	ProviderID     *string `gorm:"type:varchar(255)"`
	ErrorMessage   *string `gorm:"type:text"`
}

func New(recipient string, channel Channel, content string, priority Priority) *Notification {
	now := time.Now().UTC()
	return &Notification{
		ID:        uuid.NewString(),
		Recipient: recipient,
		Channel:   channel,
		Content:   content,
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (n *Notification) CanCancel() bool {
	return n.Status == StatusPending || n.Status == StatusQueued
}

func (n *Notification) Cancel() error {
	if !n.CanCancel() {
		return ErrCannotCancel
	}
	n.Status = StatusCancelled
	n.UpdatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) MarkQueued() {
	n.Status = StatusQueued
	n.UpdatedAt = time.Now().UTC()
}

func (n *Notification) MarkProcessing() {
	n.Status = StatusProcessing
	n.UpdatedAt = time.Now().UTC()
}

func (n *Notification) MarkSent(providerID string) {
	now := time.Now().UTC()
	n.Status = StatusSent
	n.ProviderID = &providerID
	n.SentAt = &now
	n.UpdatedAt = now
}

func (n *Notification) MarkFailed(errMsg string) {
	now := time.Now().UTC()
	n.Status = StatusFailed
	n.ErrorMessage = &errMsg
	n.FailedAt = &now
	n.RetryCount++
	n.UpdatedAt = now
}

type Filter struct {
	Status    *Status
	Channel   *Channel
	BatchID   *string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}
