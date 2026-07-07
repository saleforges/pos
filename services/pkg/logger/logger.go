package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

var l = slog.Default()

type Config struct {
	Level  string
	Format string
}

func Init(cfg Config) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	l = slog.New(handler)
}

func Debug(msg string, args ...any) { l.Debug(msg, args...) }
func Info(msg string, args ...any)  { l.Info(msg, args...) }
func Warn(msg string, args ...any)  { l.Warn(msg, args...) }
func Error(msg string, args ...any) { l.Error(msg, args...) }

func WithCtx(ctx context.Context) *slog.Logger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return l.With(
			"trace_id", span.SpanContext().TraceID().String(),
			"span_id", span.SpanContext().SpanID().String(),
		)
	}
	return l
}

func With(args ...any) *slog.Logger {
	return l.With(args...)
}
