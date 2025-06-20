package gin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	LabelHTTPQuery    = "http.query"
	LabelHTTPMethod   = "http.method"
)

var (
	method        = attribute.Key("method")
	correlationID = attribute.Key("correlation.id")
	optionshttp   = []otelhttptrace.Option{otelhttptrace.WithPropagators(commontrace.Propagator)}
	//	DefaultTelemetryBucketBoundaries 100ms 200ms 400ms 800ms 1s 2s 4s 8s 16s 32s 1m
	DefaultTelemetryBucketBoundaries = []float64{
		100,
		500,
		float64(time.Second.Milliseconds() * 1),
		float64(time.Second.Milliseconds() * 2),
		float64(time.Second.Milliseconds() * 3),
		float64(time.Second.Milliseconds() * 4),
	}
	binaryContentTypes = []string{
		"application/octet-stream",
		"image/",
		"audio/",
		"video/",
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
			Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), requestMethodPath(gctx))
		defer span.End()
		// 设置 trace 属性
		span.SetAttributes(method.String(request.Method+" "+request.URL.Path),
			correlationID.String(span.SpanContext().TraceID().String()+"-"+span.SpanContext().SpanID().String()),
			semconv.ClientAddress(request.Header.Get(xForwardedFor)),
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(env),
			semconv.UserAgentOriginal(request.Header.Get(userAgent)))

		logRequest(ctx, request, logger)
		gctx.Request = gctx.Request.Clone(ctx)
		gctx.Writer = &httpResponseLogger{
			ResponseWriter: gctx.Writer,
			serviceName:    serviceName,
			env:            env,
			ginCtx:         gctx,
			context:        ctx,
			logger:         logger,
			httpRequest:    request,
			startTime:      time.Now(),
			span:           span,
		}
		// attach trace id
		gctx.Writer.Header().Add("trace-id", span.SpanContext().TraceID().String())
		gctx.Next()

	}
}

func logRequest(ctx context.Context, r *http.Request, logger logger.Logger) {
	path := r.URL.Path
	header := r.Header
	var requestBody string
	if !isBinaryRequest(r) && !isMultipartRequest(r) {
		rawBody, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewBuffer(rawBody))
			requestBody = string(rawBody)
		}
	}
	logger.Info(ctx, "Http Request",
		zap.String(LabelHTTPService, path),
		zap.String(LabelHTTPQuery, r.URL.RawQuery),
		zap.Any(LabelHTTPHeader, header),
		zap.Any(LabelHTTPRequest, requestBody),
		zap.Any(LabelHTTPMethod, r.Method),
	)
}

func isMultipartRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "multipart/")
}

func isBinaryRequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	// Check if the Content-Type is in the list of binary content types
	for _, binaryType := range binaryContentTypes {
		if strings.HasPrefix(contentType, binaryType) {
			return true
		}
	}
	return false
}

type httpResponseLogger struct {
	gin.ResponseWriter
	serviceName string
	env         string
	ginCtx      *gin.Context
	context     context.Context
	logger      logger.Logger
	httpRequest *http.Request
	startTime   time.Time
	span        trace.Span
}

func (hrl *httpResponseLogger) Header() http.Header {
	return hrl.ResponseWriter.Header()
}

func (hrl *httpResponseLogger) Write(bytes []byte) (int, error) {
	// add metric component
	request := hrl.httpRequest
	ctx := hrl.context
	n, err := hrl.ResponseWriter.Write(bytes)
	if err != nil {
		return 0, err
	}
	hrl.logger.Info(hrl.context, "Http Response",
		zap.String(LabelHTTPService, request.URL.Path),
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
		path := requestMethodPath(hrl.ginCtx)
		attrs := []attribute.KeyValue{
			semconv.HTTPRoute(path),
			semconv.HTTPRequestMethodKey.String(request.Method),
			semconv.DeploymentEnvironmentKey.String(hrl.env),
			semconv.HTTPResponseStatusCodeKey.Int(hrl.Status()),
			// comment: because will record too many attribute. body size is Non-enumerable
			//semconv.HTTPRequestBodySize(int(request.ContentLength)),
			//semconv.HTTPResponseBodySize(hrl.ginCtx.Writer.Size()),
		}
		hrl.span.SetAttributes(attrs...)

		if counter, err := meter.Int64Counter(metricName); err == nil {
			counter.Add(ctx, 1, metric.WithAttributes(attrs...))
		}

		if histogram, err := meter.Int64Histogram(metricName+".histogram",
			metric.WithExplicitBucketBoundaries(DefaultTelemetryBucketBoundaries...)); err == nil {
			histogram.Record(ctx, time.Since(startTime).Milliseconds(),
				metric.WithAttributes(attrs...))
		}

	}(hrl.startTime)

	return n, nil
}

func requestMethodPath(ctx *gin.Context) string {
	path := ctx.Request.URL.Path
	if matchedTemplatePath := ctx.FullPath(); matchedTemplatePath != "" {
		path = matchedTemplatePath
	}
	return path
}

func (hrl *httpResponseLogger) WriteHeader(statusCode int) {
	hrl.ResponseWriter.WriteHeader(statusCode)
}
