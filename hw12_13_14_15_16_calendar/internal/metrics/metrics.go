package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP метрики.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "calendar_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Бизнес-метрики для событий календаря.
	EventsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_created_total",
			Help: "Total number of events created",
		},
	)

	EventsUpdatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_updated_total",
			Help: "Total number of events updated",
		},
	)

	EventsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_deleted_total",
			Help: "Total number of events deleted",
		},
	)

	EventsRetrievedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_retrieved_total",
			Help: "Total number of events retrieved",
		},
	)

	EventsListedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_events_listed_total",
			Help: "Total number of events listed",
		},
		[]string{"period"},
	)

	// Метрики для scheduler.
	NotificationsSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_notifications_sent_total",
			Help: "Total number of notifications sent to queue",
		},
	)

	NotificationsSendErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_notifications_send_errors_total",
			Help: "Total number of errors when sending notifications",
		},
	)

	EventsScannedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_events_scanned_total",
			Help: "Total number of events scanned for notifications",
		},
	)

	OldEventsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_old_events_deleted_total",
			Help: "Total number of old events deleted",
		},
	)

	// Метрики для storer.
	NotificationsReceivedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_notifications_received_total",
			Help: "Total number of notifications received from queue",
		},
	)

	NotificationsSavedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_notifications_saved_total",
			Help: "Total number of notifications saved to database",
		},
	)

	NotificationsSaveErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "calendar_notifications_save_errors_total",
			Help: "Total number of errors when saving notifications",
		},
	)

	// Метрики ошибок.
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type"},
	)
)
