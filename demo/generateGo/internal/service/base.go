package service

import (
	"context"
	"generatego/internal/httpserver/middleware"
	"generatego/pkg/constant"
	"generatego/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaseService struct {
	Logger *zap.Logger
}

/*
service 接收的应该是带有超时的 context.Context. 不再传 gin.Context
*/
func (b *BaseService) CtxLog(c context.Context) *zap.Logger {
	if traceId, ok := c.Value(constant.TraceName).(string); ok && traceId != "" {
		return b.Logger.With(zap.String("traceId", traceId))
	}

	// 没有再到gin.Context 拿
	if gc, ok := c.(*gin.Context); ok {
		ctxTraceId := gc.GetString(constant.TraceName)
		if ctxTraceId != "" {
			return b.Logger.With(zap.String("traceId", ctxTraceId))
		}
	}
	return b.Logger
}

/*
service 接收的应该是带有超时的 context.Context. 不再传 gin.Context
*/
func (b *BaseService) Userinfo(c context.Context) (userInfo jwt.UserInfo) {

	userInfo, ok := c.Value(middleware.AuthUserKey).(jwt.UserInfo)
	if ok {
		return
	}

	// 没有再到 gin.Context 获取
	gc, ok := c.(*gin.Context)
	if !ok {
		return
	}
	auth, _ := gc.Get(middleware.AuthUserKey)
	userInfo, ok = auth.(jwt.UserInfo)
	if ok {
		return userInfo
	}
	b.CtxLog(gc).Error("token 获取用户信息失败")
	return
}
