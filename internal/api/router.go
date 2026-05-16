package api

import (
	"study-case/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"study-case/internal/api/handler"
	"study-case/internal/service"
)

func NewNotificationRouter(notifSvc service.Service, log zerolog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Correlation(log))
	r.Use(middleware.Metrics())

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

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}
