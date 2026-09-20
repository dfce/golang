package user

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler}
}

func (r *Router) RegisterRoute(rg *gin.RouterGroup) {
	route := rg.Group("/user")

	route.POST("", r.handler.Create)
}
