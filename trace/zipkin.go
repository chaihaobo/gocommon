package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type (
	ZipkinTracer interface {
		Tracer
		Provider() TracerProvider
		Close(ctx context.Context) error
	}

	zipkinTracer struct {
		Tracer
		exporter *zipkin.Exporter
		provider TracerProvider
	}
)

func (z *zipkinTracer) Provider() TracerProvider {
	return z.provider
}

func (z *zipkinTracer) Close(ctx context.Context) error {
	return z.exporter.Shutdown(ctx)
}

func NewZipkinTracer(config Config) (ZipkinTracer, error) {
	exporter, err := zipkin.New(config.CollectorURL)
	if err != nil {
		return nil, err
	}
	res := resource.Default()
	if svcName := config.ServiceName; svcName != "" {
		res, _ = resource.Merge(res, resource.NewWithAttributes(res.SchemaURL(), semconv.ServiceName(svcName)))
	}
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter), trace.WithResource(res))
	otel.SetTracerProvider(tp)
	return &zipkinTracer{
		Tracer:   otel.Tracer(DefaultTracerName),
		exporter: exporter,
		provider: tp,
	}, nil
}
