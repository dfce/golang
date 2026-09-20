package middleware

import (
	"context"
	"errors"
	"go-module/internal/platform/logging"
	"go-module/pkg/apperror"
	"go-module/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Errorhandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		// 客户端主动断开，不再尝试写响应
		if errors.Is(err, context.Canceled) && errors.Is(c.Request.Context().Err(), context.Canceled) {
			return
		}

		fields := logging.KV2Fields(
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"error", err,
		)
		var appErr *apperror.Error
		if errors.As(err, &appErr) && appErr.Status < http.StatusInternalServerError {
			logger.Warn("request failed", fields...)
		} else {
			logger.Error("request failed", fields...)
		}

		response.WriteError(c, err)
	}
}
