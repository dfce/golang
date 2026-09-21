package user

import (
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	svc    *UserSvc
	router *Router
}

func NewModule(db *gorm.DB, redis *datastore.RedisSvc, logger *zap.Logger, tokenService *jwt.Service) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, logger, tokenService, redis)
	handler := NewHandler(svc)
	router := NewRouter(handler, redis, tokenService)

	return &Module{svc: svc, router: router}
}

func (m *Module) Service() *UserSvc { return m.svc }

func (m *Module) RegisterRoute(rg *gin.RouterGroup) {
	m.router.RegisterRoute(rg)
}
