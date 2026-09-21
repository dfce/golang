package user

import (
	"go-module/internal/middleware"
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type Router struct {
	handler      *Handler
	redisSvc     *datastore.RedisSvc
	tokenService *jwt.Service
}

func NewRouter(handler *Handler, redisSvc *datastore.RedisSvc, tokenService *jwt.Service) *Router {
	return &Router{handler, redisSvc, tokenService}
}

func (r *Router) RegisterRoute(rg *gin.RouterGroup) {
	route := rg.Group("/user")

	route.POST("", r.handler.Create)
	route.POST("/login", r.handler.Login)

	// authMiddleware
	// route.Use(middleware.Auth(r.redisSvc, r.tokenService))
	auth := route.Group("")
	auth.Use(middleware.Auth(r.redisSvc, r.tokenService))
	auth.GET("/info", r.handler.Info)

}
