package metric

import "github.com/prometheus/client_golang/prometheus"

// Config define the configuration for Prometheus metric.
type Config struct {
	ServiceName string
	// Port to expose the metrics and listen the service.
	// if not set Port. will not listen the metric service
	Port int
	// Registry to register the metrics
	// If not set, the default registry will be used
	Registerer prometheus.Registerer
	// Gatherer to gather the metrics
	// If not set, the default gatherer will be used
	Gatherer prometheus.Gatherer
}

func (c Config) GetRegister() prometheus.Registerer {
	if c.Registerer == nil {
		return prometheus.DefaultRegisterer
	}
	return c.Registerer
}

func (c Config) GetGatherer() prometheus.Gatherer {
	if c.Gatherer == nil {
		return prometheus.DefaultGatherer
	}
	return c.Gatherer
}
