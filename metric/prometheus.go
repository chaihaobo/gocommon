package metric

import (
	"context"
	"fmt"
	"net/http"

	goprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type (
	PrometheusMetric interface {
		Meter
		MeterProvider() MeterProvider
		Register() goprometheus.Registerer
		Close(ctx context.Context) error
	}

	prometheusMetric struct {
		Meter
		exporter      *prometheus.Exporter
		meterProvider MeterProvider
		register      goprometheus.Registerer
	}
)

func (p *prometheusMetric) MeterProvider() MeterProvider {
	return p.meterProvider
}

func (p *prometheusMetric) Register() goprometheus.Registerer {
	return p.register
}

func (p *prometheusMetric) Close(ctx context.Context) error {
	return p.exporter.Shutdown(ctx)
}

// NewPrometheusMetric create a new prometheus metric
//
//	will start the prometheus server base config port.
//	Return PrometheusMetric can get the MeterProvider and Meter
//	and you can get the Registerer to register your custom metric
//	if you want get the Meter with Global scope, you can use
//
//	otel.GetMeterProvider().Meter()
func NewPrometheusMetric(config Config) (PrometheusMetric, error) {
	register := config.GetRegister()
	gatherer := config.GetGatherer()
	exp, err := prometheus.New(
		prometheus.WithRegisterer(config.GetRegister()),
	)
	if err != nil {
		return nil, err
	}
	res := resource.Default()
	if svcName := config.ServiceName; svcName != "" {
		res, _ = resource.Merge(res, resource.NewWithAttributes(res.SchemaURL(), semconv.ServiceName(svcName)))
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp), sdkmetric.WithResource(res))
	otel.SetMeterProvider(provider)
	if port := config.Port; port > 0 {
		listenMetricServer(port, register, gatherer)
	}

	return &prometheusMetric{
		Meter:         provider.Meter(DefaultMeterName),
		meterProvider: provider,
		register:      register,
		exporter:      exp,
	}, nil
}

func listenMetricServer(port int, register goprometheus.Registerer, gatherer goprometheus.Gatherer) {
	addr := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(register, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})))

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("prometheus metric server panic: %v \n", r)
			}
		}()
		fmt.Println("prometheus metric server listen on: " + addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Println("prometheus metric server start failed: " + err.Error())
		}
	}()
}
