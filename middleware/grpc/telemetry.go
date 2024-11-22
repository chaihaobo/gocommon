package grpc

import (
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	otelcontrib "go.opentelemetry.io/contrib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/chaihaobo/gocommon/logger"
	commonmetric "github.com/chaihaobo/gocommon/metric"
	commontrace "github.com/chaihaobo/gocommon/trace"
)

var (
	options         = []otelgrpc.Option{otelgrpc.WithPropagators(commontrace.Propagator)}
	messageSent     = messageType(semconv.MessageTypeSent)
	messageReceived = messageType(semconv.MessageTypeReceived)
)

var (
	DefaultTelemetryBucketBoundaries = []float64{
		100,
		500,
		float64(time.Second.Milliseconds() * 1),
		float64(time.Second.Milliseconds() * 2),
		float64(time.Second.Milliseconds() * 3),
		float64(time.Second.Milliseconds() * 4),
	}
)

const (
	LabelGRPCHeader        = "rpc.header"
	LabelGRPCRequest       = "rpc.request"
	LabelGRPCResponse      = "rpc.response"
	LabelGRPCService       = "rpc.service"
	LabelAppResponseStatus = "app.response_status"
)

const (
	clientClosedState byte = 1 << iota
	receiveEndedState
)

const (
	closeEvent streamEventType = iota
	receiveEndEvent
	errorEvent
)

type (
	messageType attribute.KeyValue

	// clientStream  wraps around the embedded grpc.ClientStream, and intercepts the RecvMsg and
	// SendMsg method call.
	clientStream struct {
		grpc.ClientStream

		desc       *grpc.StreamDesc
		events     chan streamEvent
		eventsDone chan struct{}
		finished   chan error

		receivedMessageID int
		sentMessageID     int
	}
	streamEventType int

	streamEvent struct {
		Type streamEventType
		Err  error
	}

	// serverStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
	// SendMsg method call.
	serverStream struct {
		grpc.ServerStream
		ctx context.Context

		receivedMessageID int
		sentMessageID     int
	}
)

// Event adds an event of the messageType to the span associated with the
// passed context with id and size (if message is a proto message).
func (m messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent("message", trace.WithAttributes(
			attribute.KeyValue(m),
			semconv.MessageIDKey.Int(id),
			semconv.MessageCompressedSizeKey.Int(proto.Size(p)),
		),
		)
	} else {
		span.AddEvent("message",
			trace.WithAttributes(
				attribute.KeyValue(m),
				semconv.MessageIDKey.Int(id),
			),
		)
	}
}

// TelemetryStreamClientInterceptor returns a grpc.StreamClientInterceptor suitable
// for use in a grpc.Dial call.
func TelemetryStreamClientInterceptor(serviceName, env string, logger logger.Logger) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// TODO: ask why incoming metadata need to be propagated
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		outgoingMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		metadataCopy = metadata.Join(metadataCopy, outgoingMetadata)

		tracer := otel.GetTracerProvider().Tracer(commontrace.DefaultTracerName, trace.WithInstrumentationVersion(otelcontrib.Version()))

		name, attr := spanInfo(method, cc.Target())
		var span trace.Span
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attr...),
		)

		otelgrpc.Inject(ctx, &metadataCopy, options...)
		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		startTime := time.Now()

		s, err := streamer(ctx, desc, cc, method, callOpts...)
		stream := wrapClientStream(s, desc)

		go func() {
			if err == nil {
				err = <-stream.finished
			}

			if err != nil {
				s, _ := status.FromError(err)
				span.SetStatus(codes.Error, s.Message())
			}

			span.End()
		}()

		isInboundTraffic := false
		pushAdditional(ctx, logger, name, env, method, err, nil, nil, startTime, isInboundTraffic)
		return stream, err
	}
}

func (w *clientStream) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-w.eventsDone:
	case w.events <- streamEvent{Type: eventType, Err: err}:
	}
}

func (w *clientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	switch {
	case err == nil && !w.desc.ServerStreams:
		w.sendStreamEvent(receiveEndEvent, nil)
	case err == io.EOF:
		w.sendStreamEvent(receiveEndEvent, nil)
	case err != nil:
		w.sendStreamEvent(errorEvent, err)
	default:
		w.receivedMessageID++
		messageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *clientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)

	w.sentMessageID++
	messageSent.Event(w.Context(), w.sentMessageID, m)

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (w *clientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return md, err
}

func (w *clientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()

	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	} else {
		w.sendStreamEvent(closeEvent, nil)
	}

	return err
}

func wrapClientStream(s grpc.ClientStream, desc *grpc.StreamDesc) *clientStream {
	events := make(chan streamEvent)
	eventsDone := make(chan struct{})
	finished := make(chan error)

	go func() {
		defer close(eventsDone)

		// Both streams have to be closed
		state := byte(0)

		for event := range events {
			switch event.Type {
			case closeEvent:
				state |= clientClosedState
			case receiveEndEvent:
				state |= receiveEndedState
			case errorEvent:
				finished <- event.Err
				return
			}

			if state == clientClosedState|receiveEndedState {
				finished <- nil
				return
			}
		}
	}()

	return &clientStream{
		ClientStream: s,
		desc:         desc,
		events:       events,
		eventsDone:   eventsDone,
		finished:     finished,
	}
}

// TelemetryUnaryClientInterceptor returns a grpc.UnaryClientInterceptor suitable
// for use in a grpc.Dial call.
func TelemetryUnaryClientInterceptor(serviceName, env string, logger logger.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		outgoingMetadata, _ := metadata.FromOutgoingContext(ctx)

		metadataCopy := requestMetadata.Copy()
		metadataCopy = metadata.Join(metadataCopy, outgoingMetadata)

		tracer := otel.GetTracerProvider().Tracer(commontrace.DefaultTracerName, trace.WithInstrumentationVersion(otelcontrib.Version()))

		name, attr := spanInfo(method, cc.Target())
		var span trace.Span
		ctx, span = tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		otelgrpc.Inject(ctx, &metadataCopy, options...)
		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		startTime := time.Now()

		messageSent.Event(ctx, 1, req)

		err := invoker(ctx, method, req, reply, cc, callOpts...)

		messageReceived.Event(ctx, 1, reply)

		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		}

		isInboundTraffic := false
		pushAdditional(ctx, logger, serviceName, env, method, err, req, reply, startTime, isInboundTraffic)
		return err
	}
}

// TelemetryUnaryServerInterceptor returns a grpc.UnaryServerInterceptor suitable
// for use in a grpc.NewServer call.
func TelemetryUnaryServerInterceptor(serviceName, env string, logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		bagg, spanCtx := otelgrpc.Extract(ctx, &metadataCopy, options...)
		ctx = baggage.ContextWithBaggage(ctx, bagg)

		tracer := otel.GetTracerProvider().Tracer(commontrace.DefaultTracerName, trace.WithInstrumentationVersion(otelcontrib.Version()))
		name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
		ctx, span := tracer.Start(
			trace.ContextWithRemoteSpanContext(ctx, spanCtx),
			name,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		startTime := time.Now()

		messageReceived.Event(ctx, 1, req)

		resp, err := handler(ctx, req)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			messageSent.Event(ctx, 1, s.Proto())
		} else {
			messageSent.Event(ctx, 1, resp)
		}

		isInboundTraffic := true
		pushAdditional(ctx, logger, serviceName, env, info.FullMethod, err, req, resp, startTime, isInboundTraffic)
		return resp, err
	}
}

// TelemetryStreamServerInterceptor returns a grpc.StreamServerInterceptor suitable
// for use in a grpc.NewServer call.
func TelemetryStreamServerInterceptor(serviceName, env string, logger logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		bagg, spanCtx := otelgrpc.Extract(ctx, &metadataCopy, options...)
		ctx = baggage.ContextWithBaggage(ctx, bagg)

		tracer := otel.GetTracerProvider().Tracer(commontrace.DefaultTracerName, trace.WithInstrumentationVersion(otelcontrib.Version()))
		name, attr := spanInfo(info.FullMethod, peerFromCtx(ctx))
		ctx, span := tracer.Start(
			trace.ContextWithRemoteSpanContext(ctx, spanCtx),
			name,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attr...),
		)
		defer span.End()

		startTime := time.Now()

		err := handler(srv, wrapServerStream(ctx, ss))
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		}
		isInboundTraffic := true
		pushAdditional(ctx, logger, serviceName, env, info.FullMethod, err, nil, nil, startTime, isInboundTraffic)
		return err
	}
}

func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func (w *serverStream) Context() context.Context {
	return w.ctx
}

func (w *serverStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	if err == nil {
		w.receivedMessageID++
		messageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *serverStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)

	w.sentMessageID++
	messageSent.Event(w.Context(), w.sentMessageID, m)

	return err
}

func pushAdditional(ctx context.Context,
	logger logger.Logger, serviceName, env, method string, err error,
	req interface{}, resp interface{},
	start time.Time,
	isInbound bool) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	stat := uint32(status.Code(err))
	elapsedTime := time.Since(start).Milliseconds()
	attrs := []attribute.KeyValue{
		semconv.RPCMethod(method),
		semconv.DeploymentEnvironmentKey.String(env),
		attribute.Int("grpc_status", int(stat)),
		attribute.String("error_message", errorMessage),
	}
	metricsName := serviceName
	if !isInbound {
		metricsName = metricsName + ".external_grpc"
	}

	meter := otel.Meter(commonmetric.DefaultMeterName)
	if counter, err := meter.Int64Counter(serviceName); err == nil {
		counter.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	if histogram, err := meter.Int64Histogram(metricsName+".histogram",
		metric.WithExplicitBucketBoundaries(DefaultTelemetryBucketBoundaries...)); err == nil {
		histogram.Record(ctx, elapsedTime)
	}
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
	logger.Info(ctx, "",
		zap.Any(LabelGRPCHeader, metadataCopy),
		zap.Any(LabelGRPCRequest, req),
		zap.Any(LabelGRPCResponse, resp),
		zap.String(LabelGRPCService, method),
		zap.Int(LabelAppResponseStatus, int(stat)),
	)
}

// peerFromCtx returns a peer address from a context, if one exists.
func peerFromCtx(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	return p.Addr.String()
}

// spanInfo returns a span name and all appropriate attributes from the gRPC
// method and peer address.
func spanInfo(fullMethod, peerAddress string) (string, []attribute.KeyValue) {
	attrs := []attribute.KeyValue{semconv.RPCSystemGRPC}
	name, mAttrs := parseFullMethod(fullMethod)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(peerAddress)...)
	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetworkPeerAddressKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

// parseFullMethod returns a span name following the OpenTelemetry semantic
// conventions as well as all applicable span label.KeyValue attributes based
// on a gRPC's FullMethod.
func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	lenParts := 2
	if len(parts) != lenParts {
		// Invalid format, does not follow `/package.service/method`.
		return name, []attribute.KeyValue(nil)
	}

	var attrs []attribute.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}
	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}
	return name, attrs
}
