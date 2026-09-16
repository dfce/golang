package service

import (
	"context"

	"generatego/pkg/constant"
	"generatego/pkg/jwt"

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

func (b *BaseService) Userinfo(ctx context.Context) (jwt.UserInfo, bool) {
	userInfo, ok := ctx.Value(constant.AuthUserKey).(jwt.UserInfo)
	if !ok {
		return jwt.UserInfo{}, false
	}
	return userInfo, true
}
