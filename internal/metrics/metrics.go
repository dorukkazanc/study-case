package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	NotificationsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_created_total",
			Help: "Total notifications created",
		},
		[]string{"channel", "priority"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	QueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notification_queue_depth",
			Help: "Current number of notifications in queue per channel",
		},
		[]string{"channel"},
	)

	NotificationsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_sent_total",
			Help: "Total notifications successfully sent",
		},
		[]string{"channel"},
	)

	NotificationsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_failed_total",
			Help: "Total notifications moved to dead letter queue",
		},
		[]string{"channel"},
	)
)

func init() {
	prometheus.MustRegister(NotificationsCreated, HTTPRequestDuration, QueueDepth, NotificationsSent, NotificationsFailed)
}
