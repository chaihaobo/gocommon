package metric

import "go.opentelemetry.io/otel/metric/noop"

func NewNoopMeter() Meter {
	return noop.Meter{}
}
