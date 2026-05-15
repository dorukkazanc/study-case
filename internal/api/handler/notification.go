package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	svc "study-case/internal/service/notification"
)

type NotificationHandler struct {
	svc svc.Service
}

func NewNotificationHandler(svc svc.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

type createRequest struct {
	Recipient      string  `json:"recipient"       binding:"required"`
	Channel        string  `json:"channel"         binding:"required"`
	Content        string  `json:"content"         binding:"required"`
	Priority       string  `json:"priority"`
	IdempotencyKey *string `json:"idempotency_key"`
	ScheduledAt    *string `json:"scheduled_at"`
}

type createBatchRequest struct {
	Notifications []createRequest `json:"notifications" binding:"required,min=1,max=1000"`
}

type listQuery struct {
	Status   string `form:"status"`
	Channel  string `form:"channel"`
	BatchID  string `form:"batch_id"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

func (h *NotificationHandler) Create(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *NotificationHandler) CreateBatch(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *NotificationHandler) GetByID(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *NotificationHandler) List(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *NotificationHandler) Cancel(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}

func (h *NotificationHandler) GetBatch(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
