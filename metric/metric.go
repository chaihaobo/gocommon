package metric

import "go.opentelemetry.io/otel/metric"

const (
	DefaultMeterName = "meter_default"
)

type (
	MeterProvider = metric.MeterProvider
	MeterOption   = metric.MeterOption
	Meter         = metric.Meter
)
