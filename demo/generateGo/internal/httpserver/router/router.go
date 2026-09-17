package router

import (
	"generatego/internal/config"
	"generatego/internal/health"
	"generatego/internal/httpserver/handler"
	"generatego/internal/httpserver/middleware"
	demorouter "generatego/internal/httpserver/router/demo"
	userrouter "generatego/internal/httpserver/router/user"
	"generatego/internal/port"
	"generatego/internal/service"
	"generatego/pkg/util"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "generatego/swdocs"
)

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	services *service.Registry,
	sessions port.SessionStore,
	tokens port.TokenService,
	readiness *health.Service,
) *gin.Engine {
	// gin.SetMode(gin.DebugMode) // 默认开启？？？
	// 生产/测试环境：关闭控制台 debug 输出
	isDev := util.IsDev(cfg.App.ENV)
	if !isDev {
		gin.SetMode(gin.ReleaseMode)
	}

	/*
		创建一个不带任何默认中间件的 Gin Engine
		router := gin.Default() // 已经默认开启了 Logger Recovery
	*/
	router := gin.New()

	if isDev {
		// 注册 Swagger 路由访问路径 /swagger/index.html
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	/*
		加载全局中间件
	*/
	router.Use(
		middleware.TraceID(),
		middleware.Timeout(cfg.HTTP.RequestTimeout),
		gin.CustomRecovery(middleware.Recover(logger)),
		middleware.RequestInfo(logger, cfg),
	)
	// 非生产环境才开启 自带控制台请求中间件
	if cfg.App.ENV != "prod" {
		router.Use(gin.Logger())
	}

	healthHandler := handler.NewHealthHandler(readiness)
	router.GET("/livez", healthHandler.Liveness)
	router.GET("/readyz", healthHandler.Readiness)

	// Business routes are registered explicitly so the complete route surface
	// is visible from this composition root.
	userrouter.Register(router, services.User, logger, sessions, tokens, cfg.Auth.Enabled)
	demorouter.Register(router, services.Demo, logger, sessions, tokens, cfg.Auth.Enabled)

	return router
}
