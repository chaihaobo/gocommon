package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type zapLogger struct {
	logger *zap.Logger
}

func (z *zapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx))
	z.logger.Info(msg, fields...)
}

func (z *zapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx))
	z.logger.Warn(msg, fields...)
}

func (z *zapLogger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zapCorrelationID(ctx), zap.Error(err))
	z.logger.Error(msg, fields...)
}

func zapCorrelationID(ctx context.Context) zap.Field {
	ID := correlationIDFromContext(ctx)
	return zap.String(defaultCorrelationIDLabel, ID)
}

func correlationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.HasTraceID() {
		return sc.TraceID().String() + "-" + sc.SpanID().String()
	}
	return ""
}
