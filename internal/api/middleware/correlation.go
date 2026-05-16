package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func Correlation(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqId := c.GetHeader("X-Request-ID")

		if reqId == "" {
			reqId = uuid.NewString()
		}

		c.Set("request_id", reqId)
		c.Header("X-Request-ID", reqId)
		c.Next()

		log.Info().
			Str("request_id", reqId).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("request")
	}
}
