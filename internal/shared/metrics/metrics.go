package metrics

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Registry = prometheus.NewRegistry()

	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed, partitioned by method, path template, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Histogram of HTTP request latency in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	HTTPErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "errors_total",
			Help:      "Total number of HTTP error responses partitioned by method, path template, and error type.",
		},
		[]string{"method", "path", "error_type"},
	)

	AuthFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "auth",
			Name:      "failures_total",
			Help:      "Total number of authentication failures partitioned by failure reason.",
		},
		[]string{"reason"},
	)

	RBACDenialsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "rbac",
			Name:      "denials_total",
			Help:      "Total number of RBAC authorization denials partitioned by required permission.",
		},
		[]string{"permission"},
	)

	RateLimitRejectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "ratelimit",
			Name:      "rejections_total",
			Help:      "Total number of rate limit rejections partitioned by limit tier.",
		},
		[]string{"tier"},
	)

	CacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total number of cache hits partitioned by cache name.",
		},
		[]string{"cache"},
	)

	CacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "cache",
			Name:      "misses_total",
			Help:      "Total number of cache misses partitioned by cache name.",
		},
		[]string{"cache"},
	)
)

func init() {
	Registry.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPErrorsTotal,
		AuthFailuresTotal,
		RBACDenialsTotal,
		RateLimitRejectionsTotal,
		CacheHitsTotal,
		CacheMissesTotal,
	)
}

// RegisterDBStats registers MySQL connection pool statistics gauges with the Prometheus registry.
func RegisterDBStats(db *sql.DB) {
	if db == nil {
		return
	}

	Registry.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "app",
				Subsystem: "db",
				Name:      "connections_open",
				Help:      "The number of established connections both in use and idle.",
			},
			func() float64 { return float64(db.Stats().OpenConnections) },
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "app",
				Subsystem: "db",
				Name:      "connections_in_use",
				Help:      "The number of connections currently in use.",
			},
			func() float64 { return float64(db.Stats().InUse) },
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "app",
				Subsystem: "db",
				Name:      "connections_idle",
				Help:      "The number of idle connections.",
			},
			func() float64 { return float64(db.Stats().Idle) },
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "app",
				Subsystem: "db",
				Name:      "connections_wait_count",
				Help:      "The total number of connections waited for.",
			},
			func() float64 { return float64(db.Stats().WaitCount) },
		),
	)
}

// RecordHTTPRequest records an HTTP request completion.
func RecordHTTPRequest(method, path, status string, durationSeconds float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path, status).Observe(durationSeconds)
}

// RecordHTTPError records an HTTP error occurrence.
func RecordHTTPError(method, path, errorType string) {
	HTTPErrorsTotal.WithLabelValues(method, path, errorType).Inc()
}

// RecordAuthFailure records an authentication failure.
func RecordAuthFailure(reason string) {
	AuthFailuresTotal.WithLabelValues(reason).Inc()
}

// RecordRBACDenial records an RBAC authorization rejection.
func RecordRBACDenial(permission string) {
	RBACDenialsTotal.WithLabelValues(permission).Inc()
}

// RecordRateLimitRejection records a rate limit violation.
func RecordRateLimitRejection(tier string) {
	RateLimitRejectionsTotal.WithLabelValues(tier).Inc()
}

// RecordCacheHit records a cache hit.
func RecordCacheHit(cacheName string) {
	CacheHitsTotal.WithLabelValues(cacheName).Inc()
}

// RecordCacheMiss records a cache miss.
func RecordCacheMiss(cacheName string) {
	CacheMissesTotal.WithLabelValues(cacheName).Inc()
}

// Handler returns a Gin HandlerFunc that serves Prometheus metrics.
func Handler() gin.HandlerFunc {
	h := promhttp.HandlerFor(Registry, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
