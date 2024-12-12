package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/chaihaobo/gocommon/logger/encoder"
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
	var zp = config.ZapLogger
	var logRotate *lumberjack.Logger
	var err error
	if zp == nil {
		zp, logRotate, err = new(config)
	}
	if err != nil {
		return nil, nil, err
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

func new(config Config) (*zap.Logger, *lumberjack.Logger, error) {
	encoderMapping := map[string]func(cfg zapcore.EncoderConfig) zapcore.Encoder{
		"json":      zapcore.NewJSONEncoder,
		"console":   zapcore.NewConsoleEncoder,
		"jsoncolor": encoder.NewJSONColorEncoder,
	}
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	c.EncoderConfig.LevelKey = "severity"
	level := zapcore.DebugLevel
	if lvl, err := zapcore.ParseLevel(config.Level); err == nil {
		level = lvl
	}
	encoding := c.Encoding
	if _, ok := encoderMapping[config.Encoding]; ok {
		encoding = config.Encoding
	}
	core := zapcore.NewNopCore()
	writerSyncers := make([]zapcore.WriteSyncer, 0)
	var logRotate *lumberjack.Logger
	if !config.SkipStdOutput {
		writerSyncers = append(writerSyncers, zapcore.Lock(os.Stdout))
	}
	if config.FileName != "" {
		logRotate = &lumberjack.Logger{
			Filename:  config.FileName,
			MaxSize:   config.MaxSize,
			MaxAge:    config.MaxAge,
			LocalTime: false,
			Compress:  true,
		}
		writerSyncers = append(writerSyncers, zapcore.AddSync(logRotate))
	}
	finalSyncer := zapcore.NewMultiWriteSyncer(writerSyncers...)
	if config.BufferSize > 0 {
		flushInterval := time.Second
		if config.FlushBufferInterval > 0 {
			flushInterval = config.FlushBufferInterval
		}
		finalSyncer = &zapcore.BufferedWriteSyncer{
			WS:            finalSyncer,
			Size:          config.BufferSize,
			FlushInterval: flushInterval,
		}
	}
	core = zapcore.NewCore(encoderMapping[encoding](c.EncoderConfig), finalSyncer, level)
	var options []zap.Option
	if config.WithCaller {
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}
	zp := zap.New(core, options...)
	return zp, logRotate, nil
}
