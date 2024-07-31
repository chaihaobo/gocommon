package trace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/zipkin"
)

type (
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

func NewZipkinTracer(config Config) (CloseableTracer, error) {
	exporter, err := zipkin.New(config.CollectorURL)
	if err != nil {
		return nil, err
	}
	return newCloseableTracer(config, exporter), nil
}
