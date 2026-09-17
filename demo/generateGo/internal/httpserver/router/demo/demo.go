package demo

import (
	"fmt"
	"generatego/internal/httpserver/handler"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/port"
	"generatego/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(
	r *gin.Engine,
	demoService *service.DemoService,
	logger *zap.Logger,
	sessions port.SessionStore,
	tokens port.TokenService,
	authEnabled bool,
) {
	demoHandler := handler.NewDemoHandler(demoService, logger)

	demo := r.Group("/demo")

	demo.GET("/", demoHandler.TestGet)
	demo.GET("/test", func(c *gin.Context) {
		fmt.Println("/demo/test")
	})

	demo.POST(
		"/checkpost",
		middleware.Auth(
			sessions,
			tokens,
			middleware.WithEnabled(authEnabled),
			middleware.WithSkip(),
		),
		demoHandler.CheckPostInfo,
	)

	demo.POST("/redistest", demoHandler.RedisTest)
}
