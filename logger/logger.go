package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// This is the default label for the correlation ID field.
const defaultCorrelationIDLabel string = "_cID"

// Logger Interface. All methods SHOULD be safe for concurrent use.
type Logger interface {
	// Info logs a message at Info level
	Info(ctx context.Context, msg string, fields ...zap.Field)
	// Warn logs a message at Warn level
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	// Error logs a message at Errors level
	Error(ctx context.Context, msg string, err error, fields ...zap.Field)
}

// New create new instant for the Logger.
func New(config Config) (Logger, func() error, error) {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	c.EncoderConfig.LevelKey = "severity"

	var logRotate *lumberjack.Logger
	var zp = config.ZapLogger
	if zp == nil {
		zapLogger, err := c.Build()
		zp = zapLogger
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize zap logger %w", err)
		}
		core := zp.Core()
		if config.FileName != "" {
			logRotate = &lumberjack.Logger{
				Filename:  config.FileName,
				MaxSize:   config.MaxSize,
				MaxAge:    config.MaxAge,
				LocalTime: false,
				Compress:  true,
			}

			core = zapcore.NewTee(zp.Core(), zapcore.
				NewCore(zapcore.NewJSONEncoder(c.EncoderConfig), zapcore.AddSync(logRotate), zap.InfoLevel))

		}
		options := []zap.Option{}
		if config.WithCaller {
			options = append(options, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
		}

		zp = zap.New(core, options...)
	}

	return &zapLogger{
			logger: zp,
		}, func() (err error) {
			err = zp.Sync()
			if logRotate != nil {
				err = logRotate.Close()
			}
			return
		}, nil
}
