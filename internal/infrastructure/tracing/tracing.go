package tracing

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Config holds tracing configuration.
type Config struct {
	Endpoint    string // Jaeger OTLP HTTP endpoint (e.g. "localhost:4318")
	ServiceName string
	Enabled     bool
}

// Init initializes the OpenTelemetry tracer provider with OTLP HTTP exporter (Jaeger).
// Returns a shutdown function that must be called on application exit.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if !cfg.Enabled {
		logger.Log.Info("tracing disabled")
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Log.Info("tracing initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("service", cfg.ServiceName),
	)

	return tp.Shutdown, nil
}

// Tracer returns a named tracer for the given component.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
