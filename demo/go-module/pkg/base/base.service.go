package base

import (
	"context"
	"go-module/pkg/constant"

	"go.uber.org/zap"
)

type BaseService struct {
	Logger *zap.Logger
}

func (b *BaseService) CtxLog(ctx context.Context) *zap.Logger {
	if traceId, ok := ctx.Value(constant.TraceKey).(string); ok && traceId != "" {
		return b.Logger.With(zap.String(constant.TraceKey, traceId))
	}
	return b.Logger
}
