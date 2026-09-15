package base

import (
	"context"
	"generatego/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"generatego/pkg/constant"
)

type BaseService struct {
	Logger *zap.Logger
}

func (b *BaseService) CtxLog(c context.Context) *zap.Logger {
	if gc, ok := c.(*gin.Context); ok {
		ctxTraceId := gc.GetString(constant.TraceName)
		if ctxTraceId != "" {
			return b.Logger.With(zap.String("traceId", ctxTraceId))
		}
	}
	return b.Logger
}

func (b *BaseService) Userinfo(c context.Context) (userInfo jwt.UserInfo) {
	gc, ok := c.(*gin.Context)
	if !ok {
		return
	}
	auth, _ := gc.Get(constant.TraceName)
	userInfo, ok = auth.(jwt.UserInfo)

	if ok {
		return userInfo
	}
	b.CtxLog(gc).Error("token 获取用户信息失败")
	return
}
