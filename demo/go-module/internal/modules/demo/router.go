package demo

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	handler *DemoHandler
}

func NewDemoRouter(h *DemoHandler) *Router {
	return &Router{handler: h}
}

func (r *Router) RegisterRoute(rg *gin.RouterGroup) {

	demoRoute := rg.Group("/demo")

	demoRoute.GET("/ready", r.handler.Ready)
	demoRoute.GET("/test-jwt", r.handler.TestJWT)
	demoRoute.GET("/test-timeout", r.handler.TestTimeout)
}
