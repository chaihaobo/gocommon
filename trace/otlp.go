package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type (
	otlpTracer struct {
		Tracer
		exporter *otlptrace.Exporter
	}
)

func (o otlpTracer) Close(ctx context.Context) error {
	return o.exporter.Shutdown(ctx)
}

func NewOTLPTracer(config Config) (CloseableTracer, error) {
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpointURL(config.CollectorURL))
	if err != nil {
		return nil, err
	}
	return newCloseableTracer(config, exporter), nil
}
