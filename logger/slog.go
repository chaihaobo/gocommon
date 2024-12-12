package logger

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type traceSlogHandler struct {
	slog.Handler
}

func (t traceSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	correlationID := correlationIDFromContext(ctx)
	record.AddAttrs(slog.String(defaultCorrelationIDLabel, correlationID))
	return t.Handler.Handle(ctx, record)
}

func NewSlogHandler(config Config) (slog.Handler, func() error, error) {
	zp, logRotate, err := new(config)
	if err != nil {
		return nil, nil, err
	}
	return SlogHandlerFromZap(zp), closer(zp, logRotate), nil
}

func SlogHandlerFromZap(zp *zap.Logger) slog.Handler {
	return &traceSlogHandler{zapslog.NewHandler(zp.Core())}
}

func SetSlogDefault(handler slog.Handler) {
	slog.SetDefault(slog.New(handler))
}

func SetSlogDefaultFromLogger(logger Logger) {
	slog.SetDefault(slog.New(SlogHandlerFromLogger(logger)))
}

func SlogHandlerFromLogger(logger Logger) slog.Handler {
	zapLogger, ok := logger.(*zapLogger)
	if !ok {
		panic("logger is not a zap logger")
	}
	return SlogHandlerFromZap(zapLogger.logger)
}
