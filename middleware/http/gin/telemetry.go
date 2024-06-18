package gin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/chaihaobo/gocommon/logger"
	commonmetric "github.com/chaihaobo/gocommon/metric"
	commontrace "github.com/chaihaobo/gocommon/trace"
)

const (
	xForwardedFor = "x-forwarded-for"
	userAgent     = "user-agent"
)

const (
	LabelHTTPHeader   = "http.header"
	LabelHTTPRequest  = "http.request"
	LabelHTTPResponse = "http.response"
	LabelHTTPStatus   = "http.status"
	LabelHTTPService  = "http.service"
	LabelHTTPMethod   = "http.method"
)

var (
	method        = attribute.Key("method")
	correlationID = attribute.Key("correlation.id")
	optionshttp   = []otelhttptrace.Option{otelhttptrace.WithPropagators(commontrace.Propagator)}
	//	DefaultTelemetryBucketBoundaries 100ms 200ms 400ms 800ms 1s 2s 4s 8s 16s 32s 1m
	DefaultTelemetryBucketBoundaries = []float64{
		float64(time.Millisecond * 100),
		float64(time.Millisecond * 200),
		float64(time.Millisecond * 400),
		float64(time.Millisecond * 800),
		float64(time.Second * 1),
		float64(time.Second * 2),
		float64(time.Second * 4),
		float64(time.Second * 8),
		float64(time.Second * 16),
		float64(time.Second * 32),
		float64(time.Minute * 1),
	}
)

func TelemetryMiddleware(serviceName, env string, logger logger.Logger) gin.HandlerFunc {
	return func(gctx *gin.Context) {
		request := gctx.Request
		ctx := request.Context()
		// send tracer
		// 从 http 头中解析 trace 信息
		_, _, spanCtx := otelhttptrace.Extract(ctx, request, optionshttp...)

		// 记录本次请求的 trace 信息
		ctx, span := otel.GetTracerProvider().Tracer(commontrace.DefaultTracerName, trace.WithInstrumentationVersion(otelcontrib.Version())).
			Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), request.URL.Path)
		defer span.End()
		// 设置 trace 属性
		span.SetAttributes(method.String(request.Method+" "+request.URL.Path),
			correlationID.String(span.SpanContext().TraceID().String()+"-"+span.SpanContext().SpanID().String()),
			semconv.ClientAddress(request.Header.Get(xForwardedFor)),
			semconv.ServiceName(serviceName),
			semconv.UserAgentOriginal(request.Header.Get(userAgent)))

		logRequest(ctx, request, logger)
		gctx.Writer = &httpResponseLogger{
			ResponseWriter: gctx.Writer,
			serviceName:    serviceName,
			env:            env,
			ginCtx:         gctx,
			context:        ctx,
			logger:         logger,
			httpRequest:    request,
		}
		gctx.Next()

	}
}

func logRequest(ctx context.Context, r *http.Request, logger logger.Logger) {
	path := r.URL.Path
	header := r.Header
	var requestBody string
	rawBody, err := io.ReadAll(r.Body)
	if err == nil {
		r.Body = io.NopCloser(bytes.NewBuffer(rawBody))
		requestBody = string(rawBody)
	}

	logger.Info(ctx, "Http Request",
		zap.String(LabelHTTPService, path),
		zap.Any(LabelHTTPHeader, header),
		zap.Any(LabelHTTPRequest, requestBody),
		zap.Any(LabelHTTPMethod, r.Method),
	)
}

type httpResponseLogger struct {
	gin.ResponseWriter
	serviceName string
	env         string
	ginCtx      *gin.Context
	context     context.Context
	logger      logger.Logger
	httpRequest *http.Request
}

func (hrl *httpResponseLogger) Header() http.Header {
	return hrl.ResponseWriter.Header()
}

func (hrl *httpResponseLogger) Write(bytes []byte) (int, error) {
	// add metric component
	request := hrl.httpRequest
	path := hrl.httpRequest.URL.Path
	ctx := hrl.context
	n, err := hrl.ResponseWriter.Write(bytes)
	if err != nil {
		return 0, err
	}
	hrl.logger.Info(hrl.context, "Http Response",
		zap.String(LabelHTTPService, path),
		zap.String(LabelHTTPResponse, string(bytes)),
		zap.Int(LabelHTTPStatus, hrl.ResponseWriter.Status()),
	)

	meter := otel.Meter(commonmetric.DefaultMeterName)
	// send metric
	defer func(startTime time.Time) {
		metricName := fmt.Sprintf("%s_%s", method, request.URL.Path)
		if hrl.serviceName != "" {
			metricName = hrl.serviceName
		}
		if hrl.ginCtx.FullPath() != "" {
			path = hrl.ginCtx.FullPath()
		}
		attrs := []attribute.KeyValue{
			semconv.HTTPRoute(path),
			semconv.HTTPRequestMethodKey.String(request.Method),
			semconv.DeploymentEnvironmentKey.String(hrl.env),
			semconv.HTTPResponseStatusCodeKey.Int(hrl.Status()),
			semconv.HTTPRequestBodySize(int(request.ContentLength)),
			semconv.HTTPResponseBodySize(hrl.ginCtx.Writer.Size()),
		}
		if counter, err := meter.Int64Counter(metricName); err == nil {
			counter.Add(ctx, 1, metric.WithAttributes(attrs...))
		}

		if histogram, err := meter.Int64Histogram(metricName+".histogram",
			metric.WithExplicitBucketBoundaries(DefaultTelemetryBucketBoundaries...)); err == nil {
			histogram.Record(ctx, time.Since(startTime).Milliseconds(),
				metric.WithAttributes(attrs...))
		}

	}(time.Now())

	return n, nil
}

func (hrl *httpResponseLogger) WriteHeader(statusCode int) {
	hrl.ResponseWriter.WriteHeader(statusCode)
}
