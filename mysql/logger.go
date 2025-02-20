package mysql

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	commonLogger "github.com/chaihaobo/gocommon/logger"
)

var _ logger.Interface = (*GormLogger)(nil)

type GormLogger struct {
	logger commonLogger.Logger
}

func (g *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *g
	return &newlogger
}

func (g *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.logger.Info(ctx, msg, zap.Any("msg", msg), zap.Any("data", data))
}

func (g *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.logger.Warn(ctx, msg, zap.Any("msg", msg), zap.Any("data", data))
}

func (g *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.logger.Warn(ctx, msg, zap.Any("msg", msg), zap.Any("data", data))
}

func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		g.logger.Error(ctx, "sql execute happen error", err, zap.Any("elapsed", fmt.Sprintf("%vms", float64(elapsed.Nanoseconds())/1e6)), zap.Any("rows", rows), zap.Any("sql", sql))
	default:
		g.logger.Info(ctx, "sql execute info", zap.Any("elapsed", fmt.Sprintf("%vms", float64(elapsed.Nanoseconds())/1e6)), zap.Any("rows", rows), zap.Any("sql", sql))
	}
}

func NewGormLogger(logger commonLogger.Logger) *GormLogger { return &GormLogger{logger: logger} }
