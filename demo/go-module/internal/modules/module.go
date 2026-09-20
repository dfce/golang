package modules

import (
	"context"
	"fmt"
	"go-module/internal/config"
	"go-module/internal/middleware"
	demo "go-module/internal/modules/demo"
	user "go-module/internal/modules/user"
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"
	"go-module/pkg/response"
	"go-module/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type AppModule interface {
	RegisterRoute(rg *gin.RouterGroup)
}

// BuildAppEngine 将所有 NewModule 逻辑和依赖组装收拢于此
func BuildAppEngine(
	cfg *config.Config,
	dbs datastore.Databases,
	redis *datastore.RedisClient,
	logger *zap.Logger,
	tokenService *jwt.Service,
) *gin.Engine {
	// 单库链接直接取 primary
	db := dbs["primary"]

	// ==================================================
	// 1. 集中实例化所有业务模块并处理依赖(Assemble Modules)
	// ==================================================
	// A. 实例化原子模块
	demoModule := demo.NewModule(db, redis, logger, tokenService)
	userModule := user.NewModule(db, redis, logger, tokenService)
	// B. 如果有循环依赖/Setter对拧，也在这里处理(Setter 注入模式)
	// demoModule.SetSomeServices(orderModule.Service())
	// demoModule.SetSomeServices({orderService: orderModule.Service(), userService: userMOdule.Service()}) // 待验证
	// C. 实例化上层聚合模块(传入基础模块的标准 Service 接口)

	// ==================================================
	// 2. 初始化 Web 引擎并注册全量路由
	// ==================================================
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
	r := gin.New()

	if isDev {
		// 注册 Swagger 路由访问路径 /swagger/index.html
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	/*
		加载全局中间件
	*/
	r.Use(
		middleware.TraceID(),
		middleware.AccessLog(logger),

		gin.CustomRecovery(middleware.Recover(logger)),
		middleware.Errorhandler(logger),
		middleware.Timeout(cfg.HTTP.RequestTimeout),
		middleware.RequestInfo(logger, cfg),
	)
	// 非生产环境才开启 自带控制台请求中间件
	if cfg.App.ENV != "prod" {
		r.Use(gin.Logger())
	}

	// 全局 API 路由网关根组
	apiV1 := r.Group("/api/v1")

	// Health 用于K8S（Liveness/Readiness探测）等系统监控 健康检查
	healty(r)

	// 所有自装配完毕的模块切片
	modules := []AppModule{
		demoModule,
		userModule,
	}

	for _, module := range modules {
		module.RegisterRoute(apiV1)
	}

	logger.Info(fmt.Sprintf("🧩 全量业务模块(%d个)自装配与路由挂载完毕", len(modules)))

	return r
}

func healty(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		_, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		response.OK(c, gin.H{
			"status": "healthy",
		})
	})
}
