package redis

import (
	"context"
	"time"

	domain "study-case/internal/domain/notification"
	"study-case/internal/metrics"
)

type MetricsCollector struct {
	queue    *PriorityQueue
	interval time.Duration
}

func NewMetricsCollector(queue *PriorityQueue, interval time.Duration) *MetricsCollector {
	return &MetricsCollector{queue: queue, interval: interval}
}

func (c *MetricsCollector) Start(ctx context.Context) {
	channels := []domain.Channel{domain.ChannelEmail, domain.ChannelPush, domain.ChannelSMS}
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, ch := range channels {
				depth, err := c.queue.Depth(ctx, ch)
				if err != nil {
					continue
				}
				metrics.QueueDepth.WithLabelValues(string(ch)).Set(float64(depth))
			}
		}
	}
}
