package otel

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"github.com/saleforge/pos/services/pkg/logger"
)

func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return tracerProvider().Start(ctx, name, trace.WithAttributes(attrs...))
}

func GetTracer() trace.Tracer {
	return tracerProvider()
}

func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			start := time.Now()
			err := next(c)
			duration := time.Since(start)
			status := c.Response().Status
			logger.WithCtx(req.Context()).Info("request",
				"method", req.Method,
				"path", c.Path(),
				"status", status,
				"duration", duration.String(),
			)
			return err
		}
	}
}