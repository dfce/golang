package logging

import (
	"context"
	"errors"
	"generatego/pkg/constant"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormZapLogger 自定义 GORM 日志适配器
type GormZapLogger struct {
	zapLog        *zap.Logger
	SlowThreshold time.Duration // 慢SQL阈值
}

func NewGormZapLogger(loger *zap.Logger) *GormZapLogger {
	return &GormZapLogger{
		zapLog:        loger,
		SlowThreshold: 200 * time.Millisecond, // 超过200ms的SQL视为慢查询
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
		if val := ctx.Value(constant.TraceName); val != "" {
			if trace_id, ok := val.(string); ok {
				traceID = trace_id
			}
		}
	}

	// 组装日志输出字段
	logFields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}
	if traceID != "" {
		logFields = append(logFields, zap.String("traceID", traceID))
	}

	// 1. 如果SQL执行报错（且不是“未找到记录”这种业务正常错误），记为 ERROR日志
	if err != nil && !errors.Is(err, logger.ErrRecordNotFound) {
		errFields := append(logFields, zap.Error(err))
		l.zapLog.Error("SQL EXEC ERROR \n"+sql,
			errFields...,
		)
		return
	}

	// 2. 执行时间超过阈值，记为 WARN级别的慢SQL日志
	if elapsed > l.SlowThreshold {
		l.zapLog.Warn("SLOW SQL DETECTED \n"+sql,
			logFields...,
		)
		return
	}

	// 3. 正常情况：非生产环境打印所有执行的 SQL（DEBUG/INFO）
	if os.Getenv("ENV") != "prod" {
		l.zapLog.Info("SQL TRACE \n"+sql,
			logFields...,
		)
	}
}
