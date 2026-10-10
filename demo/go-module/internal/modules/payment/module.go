package payment

import (
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Module struct {
	svc    *PaymentSvc
	router *Router
}

func NewModule(db *gorm.DB, redis *datastore.RedisSvc, logger *zap.Logger, tokenSvc *jwt.Service, deps Dependencies) *Module {
	repo := NewRepository(db)
	svc := NewService(repo, logger, redis, deps)
	handler := NewHandler(svc)
	router := NewRouter(handler, redis, tokenSvc)

	return &Module{svc: svc, router: router}
}

func (m *Module) Service() *PaymentSvc { return m.svc }
func (m *Module) RegisterRoute(rg *gin.RouterGroup) {
	m.router.RegisterRoute(rg)
}
