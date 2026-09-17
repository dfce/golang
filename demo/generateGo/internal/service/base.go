package service

import (
	"context"

	"generatego/internal/port"
	"generatego/pkg/constant"

	"go.uber.org/zap"
)

type BaseService struct {
	Logger *zap.Logger
}

func (b *BaseService) CtxLog(ctx context.Context) *zap.Logger {
	if traceID, ok := ctx.Value(constant.TraceName).(string); ok && traceID != "" {
		return b.Logger.With(zap.String(constant.TraceName, traceID))
	}
	return b.Logger
}

func (b *BaseService) Userinfo(ctx context.Context) (port.Principal, bool) {
	userInfo, ok := ctx.Value(constant.AuthUserKey).(port.Principal)
	if !ok {
		return port.Principal{}, false
	}
	return userInfo, true
}
