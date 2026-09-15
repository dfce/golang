package router

import (
	"context"
	"generatego/internal/config"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/httpserver/router/registry"
	"generatego/internal/platform/datastore"
	"generatego/internal/service"
	"generatego/pkg/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "generatego/swdocs"

	// 解耦 相互import 如果分组各自的目录的情况
	/*
		如此才能如下， 引用子路由的包， 并只执行子包的 init 函数来实现自动注册路由表
	*/
	_ "generatego/internal/httpserver/router/demo"
	_ "generatego/internal/httpserver/router/user"
)

func NewRouter(cfg *config.Config, logger *zap.Logger, services *service.Registry, redis *datastore.RedisClient) *gin.Engine {
	// 1. 根据全局环境变量设置 Gin 的运行模式
	// if isDevelopment(cfg.App.ENV) {
	// 	// 开发环境： 开启详细调试模式
	// 	// gin.SetMode(gin.DebugMode) // 默认就开启？？？
	// } else {
	// 	// 生产/测试环境：关闭控制台 debug 输出
	// 	gin.SetMode(gin.ReleaseMode)
	// }
	if !util.IsDev(cfg.App.ENV) {
		// if cfg.App.ENV == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	/*
		创建一个不带任何默认中间件的 Gin Engine
		router := gin.Default() // 已经默认开启了 Logger Recovery
	*/
	router := gin.New()

	if util.IsDev(cfg.App.ENV) {
		// 注册 Swagger 路由访问路径 /swagger/index.html
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	/*
		加载全局中间件
	*/
	router.Use(
		middleware.TraceID(),
		middleware.Timeout(2*time.Second),
		gin.CustomRecovery(middleware.Recover(logger)),
		middleware.RequestInfo(logger, cfg),
	)
	// 非生产环境才开启 自带控制台请求中间件
	if cfg.App.ENV != "prod" {
		router.Use(gin.Logger())
	}

	// Health 用于K8S（Liveness/Readiness探测）等系统监控 健康检查
	healty(router)

	// 加载业务逻辑路由
	for _, subRouter := range registry.SubRouters {
		subRouter.Register(router, services, logger, redis)
	}

	return router
}

func healty(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		_, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		c.JSON(
			http.StatusOK,
			gin.H{
				"status": "healthy",
			},
		)
	})
}
