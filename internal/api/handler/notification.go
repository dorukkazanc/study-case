package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	domain "study-case/internal/domain/notification"
	"study-case/internal/service"
)

type NotificationHandler struct {
	svc service.Service
}

func NewNotificationHandler(svc service.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

type CreateRequest struct {
	Recipient      string  `json:"recipient"        binding:"required"`
	Channel        string  `json:"channel"          binding:"required,oneof=sms email push"  example:"sms"`
	Content        string  `json:"content"          binding:"required,max=1600"              example:"Your verification code is 123456"`
	Priority       string  `json:"priority"         binding:"omitempty,oneof=high normal low" example:"normal"`
	IdempotencyKey *string `json:"idempotency_key"`
	ScheduledAt    *string `json:"scheduled_at"     example:"2026-01-01T12:00:00Z"`
}

type CreateBatchRequest struct {
	Notifications []CreateRequest `json:"notifications" binding:"required,min=1,max=1000"`
}

type listQuery struct {
	Status    string `form:"status"`
	Channel   string `form:"channel"`
	BatchID   string `form:"batch_id"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"page_size,default=20"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// Create godoc
// @Summary      Create a notification
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      CreateRequest        true  "Notification payload"
// @Success      201      {object}  NotificationResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := domain.PriorityNormal
	if req.Priority != "" {
		priority = domain.Priority(req.Priority)
	}

	svcReq := service.CreateRequest{
		Recipient:      req.Recipient,
		Channel:        domain.Channel(req.Channel),
		Content:        req.Content,
		Priority:       priority,
		IdempotencyKey: req.IdempotencyKey,
	}

	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduled_at format"})
			return
		}
		svcReq.ScheduledAt = &t
	}

	n, err := h.svc.Create(c.Request.Context(), svcReq)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toNotificationResponse(n))
}

// CreateBatch godoc
// @Summary      Create a batch of notifications
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      CreateBatchRequest   true  "Batch payload (max 1000)"
// @Success      201      {object}  CreateBatchResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /notifications/batch [post]
func (h *NotificationHandler) CreateBatch(c *gin.Context) {
	var req CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notifications := make([]service.CreateRequest, 0, len(req.Notifications))
	for _, n := range req.Notifications {
		priority := domain.PriorityNormal
		if n.Priority != "" {
			priority = domain.Priority(n.Priority)
		}
		notifications = append(notifications, service.CreateRequest{
			Recipient:      n.Recipient,
			Channel:        domain.Channel(n.Channel),
			Content:        n.Content,
			Priority:       priority,
			IdempotencyKey: n.IdempotencyKey,
		})
	}

	result, err := h.svc.CreateBatch(c.Request.Context(), service.CreateBatchRequest{
		Notifications: notifications,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateBatchResponse{
		BatchID: result.BatchID,
		Count:   result.Count,
	})
}

// GetByID godoc
// @Summary      Get notification by ID
// @Tags         notifications
// @Produce      json
// @Param        id   path      string  true  "Notification ID"
// @Success      200  {object}  NotificationResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /notifications/{id} [get]
func (h *NotificationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	n, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toNotificationResponse(n))
}

// List godoc
// @Summary      List notifications
// @Tags         notifications
// @Produce      json
// @Param        status      query     string  false  "Filter by status (pending,queued,processing,sent,failed,cancelled)"
// @Param        channel     query     string  false  "Filter by channel (sms,email,push)"
// @Param        batch_id    query     string  false  "Filter by batch ID"
// @Param        start_date  query     string  false  "Filter from date (RFC3339)"
// @Param        end_date    query     string  false  "Filter to date (RFC3339)"
// @Param        page        query     int     false  "Page number (default 1)"
// @Param        page_size   query     int     false  "Page size (default 20)"
// @Success      200  {object}  ListResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := domain.Filter{
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	if q.StartDate != "" {
		t, err := time.Parse(time.RFC3339, q.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}
		filter.StartDate = &t
	}
	if q.EndDate != "" {
		t, err := time.Parse(time.RFC3339, q.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
			return
		}
		filter.EndDate = &t
	}
	if q.Status != "" {
		s := domain.Status(q.Status)
		filter.Status = &s
	}
	if q.Channel != "" {
		ch := domain.Channel(q.Channel)
		filter.Channel = &ch
	}
	if q.BatchID != "" {
		filter.BatchID = &q.BatchID
	}

	notifications, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]*NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		resp = append(resp, toNotificationResponse(n))
	}

	c.JSON(http.StatusOK, ListResponse{
		Notifications: resp,
		Total:         total,
		Page:          q.Page,
		PageSize:      q.PageSize,
	})
}

// Cancel godoc
// @Summary      Cancel a pending notification
// @Tags         notifications
// @Param        id   path  string  true  "Notification ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /notifications/{id} [delete]
func (h *NotificationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetBatch godoc
// @Summary      Get batch statistics
// @Tags         batches
// @Produce      json
// @Param        batchID  path      string  true  "Batch ID"
// @Success      200      {object}  BatchResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /batches/{batchID} [get]
func (h *NotificationHandler) GetBatch(c *gin.Context) {
	batchID := c.Param("batchID")

	batch, err := h.svc.GetBatch(c.Request.Context(), batchID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBatchResponse(batch))
}
