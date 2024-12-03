package metrics

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

type Metrics struct {
	requestCounter    *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	responseSizes     *prometheus.HistogramVec
	activeConnections *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		responseSizes: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response sizes in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		activeConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
			[]string{"state"},
		),
	}
}

func (m *Metrics) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Increment active connections
			m.activeConnections.WithLabelValues("active").Inc()
			defer m.activeConnections.WithLabelValues("active").Dec()

			err := next(c)

			// Record metrics after the request is processed
			duration := time.Since(start).Seconds()
			status := c.Response().Status
			path := c.Path()
			method := c.Request().Method

			m.requestCounter.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
			m.requestDuration.WithLabelValues(method, path).Observe(duration)
			m.responseSizes.WithLabelValues(method, path).Observe(float64(c.Response().Size))

			return err
		}
	}
}
