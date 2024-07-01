package logger

import (
	"context"

	"go.uber.org/zap"
)

type noopLogger struct{}

func (n noopLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
}

func (n noopLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
}

func (n noopLogger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
}

func NewNoopLogger() Logger {
	return &noopLogger{}
}
