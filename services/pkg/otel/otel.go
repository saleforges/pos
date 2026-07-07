package otel

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer atomic.Pointer[trace.Tracer]
	meter  atomic.Pointer[metric.Meter]
)

func tracerProvider() trace.Tracer {
	t := tracer.Load()
	if t != nil {
		return *t
	}
	return otel.GetTracerProvider().Tracer("")
}

func meterProvider() metric.Meter {
	m := meter.Load()
	if m != nil {
		return *m
	}
	return otel.GetMeterProvider().Meter("")
}

func setTracer(t trace.Tracer) {
	tracer.Store(&t)
}

func setMeter(m metric.Meter) {
	meter.Store(&m)
}

type Config struct {
	ServiceName  string
	Environment  string
	OtelEndpoint string
	UseGRPC      bool
	Insecure     bool
}

func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	var traceExp sdktrace.SpanExporter
	if cfg.UseGRPC {
		var opts []otlptracegrpc.Option
		if cfg.OtelEndpoint != "" {
			opts = append(opts, otlptracegrpc.WithEndpoint(cfg.OtelEndpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exp, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, err
		}
		traceExp = exp
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	setTracer(tp.Tracer(cfg.ServiceName))
	setMeter(mp.Meter(cfg.ServiceName))

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
		return mp.Shutdown(ctx)
	}, nil
}