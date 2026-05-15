package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	domain "study-case/internal/domain/notification"
)

type notificationResponse struct {
	ID             string     `json:"id"`
	BatchID        *string    `json:"batch_id,omitempty"`
	Recipient      string     `json:"recipient"`
	Channel        string     `json:"channel"`
	Content        string     `json:"content"`
	Priority       string     `json:"priority"`
	Status         string     `json:"status"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	RetryCount     int        `json:"retry_count"`
	ProviderID     *string    `json:"provider_id,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
}

type batchResponse struct {
	ID         string    `json:"id"`
	Total      int       `json:"total"`
	Pending    int       `json:"pending"`
	Queued     int       `json:"queued"`
	Processing int       `json:"processing"`
	Sent       int       `json:"sent"`
	Failed     int       `json:"failed"`
	Cancelled  int       `json:"cancelled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type listResponse struct {
	Notifications []*notificationResponse `json:"notifications"`
	Total         int                     `json:"total"`
	Page          int                     `json:"page"`
	PageSize      int                     `json:"page_size"`
}

type createBatchResponse struct {
	BatchID string `json:"batch_id"`
	Count   int    `json:"count"`
}

func toNotificationResponse(n *domain.Notification) *notificationResponse {
	return &notificationResponse{
		ID:             n.ID,
		BatchID:        n.BatchID,
		Recipient:      n.Recipient,
		Channel:        string(n.Channel),
		Content:        n.Content,
		Priority:       string(n.Priority),
		Status:         string(n.Status),
		IdempotencyKey: n.IdempotencyKey,
		ScheduledAt:    n.ScheduledAt,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
		SentAt:         n.SentAt,
		FailedAt:       n.FailedAt,
		RetryCount:     n.RetryCount,
		ProviderID:     n.ProviderID,
		ErrorMessage:   n.ErrorMessage,
	}
}

func toBatchResponse(b *domain.Batch) *batchResponse {
	return &batchResponse{
		ID:         b.ID,
		Total:      b.Total,
		Pending:    b.Pending,
		Queued:     b.Queued,
		Processing: b.Processing,
		Sent:       b.Sent,
		Failed:     b.Failed,
		Cancelled:  b.Cancelled,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrBatchNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrCannotCancel), errors.Is(err, domain.ErrDuplicateIdempotencyKey):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
