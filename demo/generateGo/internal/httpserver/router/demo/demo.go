package demo

import (
	"fmt"
	"generatego/internal/httpserver/handler"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/httpserver/router/registry"
	"generatego/internal/platform/datastore"
	"generatego/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type demoRouter struct{}

// 利用 ini， 在包被加载时， 自动挂载到全局子路由注册
func init() {
	registry.RegisterSubRouters(&demoRouter{})
}

func (d *demoRouter) Register(r *gin.Engine, services *service.Registry, logger *zap.Logger, redis *datastore.RedisClient) {
	demoHandler := handler.NewDemoHandler(services.Demo, logger)

	demo := r.Group("/demo")

	demo.GET("/health", demoHandler.Ready)

	demo.GET("/", demoHandler.TestGet)
	demo.GET("/test", func(c *gin.Context) {
		fmt.Println("/demo/test")
	})

	demo.POST("/checkpost", middleware.Auth(redis, middleware.WithSkip()), demoHandler.CheckPostInfo)

	demo.POST("/redistest", demoHandler.RedisTest)

}
