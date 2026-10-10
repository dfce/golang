package base

import (
	"context"
	"go-module/pkg/constant"
	"go-module/pkg/jwt"

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

// 获取 Auth token 解析的用户相关信息
func (b *BaseService) authInfo(ctx context.Context) (jwt.UserInfo, bool) {
	userInfo, ok := ctx.Value(constant.AuthUserKey).(jwt.UserInfo)
	if !ok {
		return jwt.UserInfo{}, false
	}
	return userInfo, true
}
