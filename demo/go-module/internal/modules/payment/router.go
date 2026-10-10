package payment

import (
	"go-module/internal/middleware"
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type Router struct {
	handler  *Handler
	redisSvc *datastore.RedisSvc
	tokenSvc *jwt.Service
}

func NewRouter(handler *Handler, redisSvc *datastore.RedisSvc, tokenSvc *jwt.Service) *Router {
	return &Router{
		handler, redisSvc, tokenSvc,
	}
}

func (r *Router) RegisterRoute(rg *gin.RouterGroup) {
	auth := rg.Group("/payment")
	auth.Use(middleware.Auth(r.redisSvc, r.tokenSvc))

	auth.POST("/deposit", r.handler.Deposit)
}
