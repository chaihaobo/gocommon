package logger

import (
	"context"
	"go.uber.org/zap/exp/zapslog"
	"log/slog"
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
	zp, logRotate, err := newZapLogger(config)
	if err != nil {
		return nil, nil, err
	}
	return &traceSlogHandler{zapslog.NewHandler(zp.Core(), nil)}, closer(zp, logRotate), nil
}
