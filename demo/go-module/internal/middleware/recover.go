package middleware

import (
	"go-module/pkg/apperror"
	"go-module/pkg/response"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover(logger *zap.Logger) gin.RecoveryFunc {
	return func(c *gin.Context, panicValue any) {
		logger.Error(
			"panic recovered",
			zap.Any("panic", panicValue),
			zap.ByteString("stack", debug.Stack()),
		)

		if !c.Writer.Written() {
			response.WriteError(c, apperror.Internal("服务器内部错误"))
		}
	}
}
