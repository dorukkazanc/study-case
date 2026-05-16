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

type createRequest struct {
	Recipient      string  `json:"recipient"        binding:"required"`
	Channel        string  `json:"channel"          binding:"required,oneof=sms email push"`
	Content        string  `json:"content"          binding:"required,max=1600"`
	Priority       string  `json:"priority"         binding:"omitempty,oneof=high normal low"`
	IdempotencyKey *string `json:"idempotency_key"`
	ScheduledAt    *string `json:"scheduled_at"`
}

type createBatchRequest struct {
	Notifications []createRequest `json:"notifications" binding:"required,min=1,max=1000"`
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

func (h *NotificationHandler) Create(c *gin.Context) {
	var req createRequest
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

func (h *NotificationHandler) CreateBatch(c *gin.Context) {
	var req createBatchRequest
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

	c.JSON(http.StatusCreated, createBatchResponse{
		BatchID: result.BatchID,
		Count:   result.Count,
	})
}

func (h *NotificationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	n, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toNotificationResponse(n))
}

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

	resp := make([]*notificationResponse, 0, len(notifications))
	for _, n := range notifications {
		resp = append(resp, toNotificationResponse(n))
	}

	c.JSON(http.StatusOK, listResponse{
		Notifications: resp,
		Total:         total,
		Page:          q.Page,
		PageSize:      q.PageSize,
	})
}

func (h *NotificationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *NotificationHandler) GetBatch(c *gin.Context) {
	batchID := c.Param("batchID")

	batch, err := h.svc.GetBatch(c.Request.Context(), batchID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toBatchResponse(batch))
}
