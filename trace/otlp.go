package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type (
	otlpTracer struct {
		Tracer
		exporter *otlptrace.Exporter
	}
)

func (o otlpTracer) Close(ctx context.Context) error {
	return o.Close(ctx)
}

func NewOTLPTracer(config Config) (CloseableTracer, error) {
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpointURL(config.CollectorURL))
	if err != nil {
		return nil, err
	}
	res := resource.Default()
	if svcName := config.ServiceName; svcName != "" {
		res, _ = resource.Merge(res, resource.NewWithAttributes(res.SchemaURL(), semconv.ServiceName(svcName)))
	}
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter), trace.WithResource(res))
	otel.SetTracerProvider(tp)
	return &otlpTracer{
		Tracer:   otel.Tracer(DefaultTracerName),
		exporter: exporter,
	}, nil
}
