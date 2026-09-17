package user

import (
	"generatego/internal/httpserver/handler"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/port"
	"generatego/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(
	r *gin.Engine,
	userService *service.UserService,
	logger *zap.Logger,
	sessions port.SessionStore,
	tokens port.TokenService,
	authEnabled bool,
) {
	userHandler := handler.NewUserHandler(userService, logger)

	user := r.Group("/user")

	user.POST("/create", userHandler.Create)
	user.POST("/login", userHandler.Login)
	user.GET(
		"/list",
		middleware.Auth(sessions, tokens, middleware.WithEnabled(authEnabled)),
		userHandler.List,
	)
}
