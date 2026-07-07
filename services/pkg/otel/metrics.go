package otel

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	HTTPRequestsInFlight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed",
	}, []string{"method", "path"})
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestDuration, HTTPRequestsInFlight)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}
			method := req.Method

			HTTPRequestsInFlight.WithLabelValues(method, path).Inc()
			defer HTTPRequestsInFlight.WithLabelValues(method, path).Dec()

			start := time.Now()
			err := next(c)
			duration := time.Since(start).Seconds()

			status := strconv.Itoa(c.Response().Status)
			HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
			HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

			return err
		}
	}
}