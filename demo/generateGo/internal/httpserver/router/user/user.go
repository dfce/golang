package user

import (
	"generatego/internal/httpserver/handler"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/httpserver/router/registry"
	"generatego/internal/platform/datastore"
	"generatego/internal/service"
	"generatego/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type userRouter struct{}

// 利用 ini， 在包被加载时， 自动挂载到全局子路由注册
func init() {
	registry.RegisterSubRouters(&userRouter{})
}

func (u *userRouter) Register(r *gin.Engine, services *service.Registry, logger *zap.Logger, redis *datastore.RedisClient, token *jwt.Service) {
	userHandler := handler.NewUserHandler(services.User, logger)

	user := r.Group("/user")

	user.POST("/create", userHandler.Create)
	user.POST("/login", userHandler.Login)
	user.GET("/list", middleware.Auth(redis, token), userHandler.List)
}
