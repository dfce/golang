package demo

import (
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	svc    *DemoService
	router *Router
}

func NewModule(db *gorm.DB, redis *datastore.RedisSvc, logger *zap.Logger, tokenService *jwt.Service) *Module {
	repo := NewDemoRepository(db, redis)
	svc := NewDemoService(repo, redis, logger, tokenService)
	handler := NewDeomHandler(svc)
	router := NewDemoRouter(handler)

	return &Module{svc: svc, router: router}
}

func (m *Module) Service() *DemoService { return m.svc }

func (m *Module) RegisterRoute(rg *gin.RouterGroup) {
	m.router.RegisterRoute(rg)
}
