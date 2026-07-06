package otel

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"github.com/saleforge/pos/services/pkg/logger"
)

func TracingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			ctx := otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			routePath := c.Path()
			if routePath == "" {
				routePath = req.URL.Path
			}

			ctx, span := tracerProvider().Start(ctx, routePath,
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.route", routePath),
					attribute.String("http.target", req.URL.Path),
					attribute.String("http.scheme", req.URL.Scheme),
					attribute.String("net.host.name", req.Host),
				),
			)
			defer span.End()

			c.SetRequest(req.WithContext(ctx))

			err := next(c)

			status := c.Response().Status
			span.SetAttributes(attribute.Int("http.status_code", status))

			if status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(status))
			} else if status >= 400 {
				span.SetStatus(codes.Error, http.StatusText(status))
			}

			return err
		}
	}
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