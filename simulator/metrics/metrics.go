package metrics

import "github.com/prometheus/client_golang/prometheus"

var Namespace = "webpushsim"

var (
	registrations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "registrations_counts",
			Help:      "Registration counts.",
		},
		[]string{"action"},
	)

	notifications = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "notifications_counts",
			Help:      "Notification counts",
		},
		[]string{"action"},
	)

	notificationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: Namespace,
			Name:      "notification_duration_seconds",
			Buckets:   prometheus.ExponentialBuckets(.1, 2, 11),
			Help:      "Seconds from notification sent to received.",
		},
	)

	connections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "connections_totals",
			Help:      "Current connection count.",
		},
	)

	connectionRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "connections_counts",
			Help:      "Connection connect and disconnect counts.",
		},
		[]string{"action"},
	)

	errs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: Namespace,
			Name:      "errors_counts",
			Help:      "Error count.",
		},
		[]string{"type"},
	)

	RegistrationSent = registrations.WithLabelValues("sent")
	RegistrationRecv = registrations.WithLabelValues("recv")
)

func init() {
	prometheus.MustRegister(registrations)

	prometheus.MustRegister(notifications)
	prometheus.MustRegister(notificationDuration)

	prometheus.MustRegister(connections)
	prometheus.MustRegister(connectionRate)

	prometheus.MustRegister(errs)
}

func ConnectionStarted() {
	connectionRate.WithLabelValues("connect").Inc()
	connections.Inc()
}

func ConnectionEnded() {
	connectionRate.WithLabelValues("disconnect").Inc()
	connections.Dec()
}

func Error(errType string) {
	errs.WithLabelValues(errType).Inc()
}

func NotificationSent(channelID string) {
	notifTimers.add(channelID)
	notifications.WithLabelValues("sent").Inc()
}

func NotificationCancel(channelID string) {
	Error("notify")
	notifTimers.del(channelID)
}

func NotificationRecv(channelID string) {
	notifTimers.finish(channelID)
	notifications.WithLabelValues("recv").Inc()
}
