package registry

import (
	"generatego/internal/platform/datastore"
	"generatego/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// subROuters 全局子路由注册表挂载点
var SubRouters []SubRouter

// 子路由[每个子路由都必须实现的标准接口]
type SubRouter interface {
	// Register 接收Gin的总引擎（或路由组），以及全量的 svcs
	// Register(gp *gin.Engine, *service.Container)
	Register(r *gin.Engine, s *service.Registry, logger *zap.Logger, redis *datastore.RedisClient)
}

func RegisterSubRouters(routers ...SubRouter) {
	SubRouters = append(SubRouters, routers...)
}
