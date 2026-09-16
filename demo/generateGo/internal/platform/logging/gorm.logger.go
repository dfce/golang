package logging

import (
	"context"
	"errors"
	"time"

	"generatego/pkg/constant"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormZapLogger 自定义 GORM 日志适配器
type GormZapLogger struct {
	zapLog        *zap.Logger
	SlowThreshold time.Duration // 慢SQL阈值
	logSQL        bool
}

func NewGormZapLogger(loger *zap.Logger, logSQL bool) *GormZapLogger {
	return &GormZapLogger{
		zapLog:        loger,
		SlowThreshold: 200 * time.Millisecond, // 超过200ms的SQL视为慢查询
		logSQL:        logSQL,
	}
}

// 实现 gorm/logger.Interface 的 4个核心方法
func (l *GormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...any) {
	l.zapLog.Sugar().Infof(msg, data...)
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.zapLog.Sugar().Warnf(msg, data...)
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...any) {
	l.zapLog.Sugar().Errorf(msg, data...)
}
func (l *GormZapLogger) Debug(ctx context.Context, msg string, data ...any) {
	l.zapLog.Sugar().Debugf(msg, data...)
}

// Trace 专门接收 GORM 的 SQL 执行事件（最核心）
func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	var traceID string
	if ctx != nil {
		traceID, _ = ctx.Value(constant.TraceName).(string)
	}

	logFields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}
	if traceID != "" {
		logFields = append(logFields, zap.String(constant.TraceName, traceID))
	}
	l.traceSQL(sql, elapsed, err, logFields)
}

func (l *GormZapLogger) traceSQL(sql string, elapsed time.Duration, err error, logFields []zap.Field) {
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		errFields := append(logFields, zap.Error(err))
		l.zapLog.Error("SQL EXEC ERROR \n"+sql,
			errFields...,
		)
		return
	}

	if elapsed > l.SlowThreshold {
		l.zapLog.Warn("SLOW SQL DETECTED \n"+sql,
			logFields...,
		)
		return
	}

	if l.logSQL {
		l.zapLog.Info("SQL TRACE \n"+sql,
			logFields...,
		)
	}
}
