package trace

import (
	"context"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
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
)
