package api

import (
	"github.com/gin-gonic/gin"

	"study-case/internal/api/handler"
	svc "study-case/internal/service/notification"
)

func NewRouter(notifSvc svc.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)

	n := handler.NewNotificationHandler(notifSvc)

	v1 := r.Group("/api/v1")
	{
		notifications := v1.Group("/notifications")
		{
			notifications.POST("", n.Create)
			notifications.POST("/batch", n.CreateBatch)
			notifications.GET("", n.List)
			notifications.GET("/:id", n.GetByID)
			notifications.DELETE("/:id", n.Cancel)
		}

		v1.GET("/batches/:batchID", n.GetBatch)
	}

	return r
}
