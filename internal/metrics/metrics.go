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
)

func init() {
	prometheus.MustRegister(NotificationsCreated, HTTPRequestDuration)
}
