package trace

import (
	"context"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	DefaultTracerName = "tracer_default"
)

var (
	Propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{},
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader)))
)

type (
	TracerProvider  = trace.TracerProvider
	Tracer          = trace.Tracer
	CloseableTracer interface {
		Tracer
		Close(ctx context.Context) error
	}

	closeableTracer struct {
		Tracer
		exporter oteltrace.SpanExporter
	}
)

func (c closeableTracer) Close(ctx context.Context) error {
	return c.exporter.Shutdown(ctx)
}

func newCloseableTracer(config Config, exporter oteltrace.SpanExporter) CloseableTracer {
	initTraceProvider(config, exporter)
	return &closeableTracer{
		Tracer:   otel.Tracer(DefaultTracerName),
		exporter: nil,
	}
}

func initTraceProvider(config Config, exporter oteltrace.SpanExporter) {
	res := resource.Default()
	if svcName := config.ServiceName; svcName != "" {
		res, _ = resource.Merge(res, resource.NewWithAttributes(res.SchemaURL(), semconv.ServiceName(svcName)))
	}
	var sampleRate = 1.0
	if config.SampleRate > 0 {
		sampleRate = config.SampleRate
	}
	sampler := oteltrace.TraceIDRatioBased(sampleRate)
	tp := oteltrace.NewTracerProvider(oteltrace.WithSyncer(exporter), oteltrace.WithResource(res), oteltrace.WithSampler(sampler))
	otel.SetTracerProvider(tp)
}
