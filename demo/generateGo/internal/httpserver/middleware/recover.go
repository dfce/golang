package middleware

import (
	"net/http"
	// "runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover(logger *zap.Logger) gin.RecoveryFunc {
	return func(c *gin.Context, err any) {

		logger.Error(
			"panic recovered",
			zap.Any("panic", err),
			// zap.ByteString("stack", debug.Stack()),
		)

		// 返回统一 JSON 错误响应
		if !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Internal Server Error, 服务器错误",
				"data":    nil,
			})
		}
	}
}
