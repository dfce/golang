package middleware

import (
	"go-module/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// 继续执行后续中间件和Handler
		c.Next()

		route := c.FullPath()
		if route == "" {
			// 404 等没有匹配路由时，FullPath可能为空
			route = c.Request.URL.Path
		}

		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("trace_id", util.TraceID(c)),
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Int("status", status),
			zap.Int("body_bytes", c.Writer.Size()),
			// zap.Duration("Execution time latency", time.Since(start)),
			zap.Int64("Execution time latency", time.Since(start).Milliseconds()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 只记录错误类型和错误信息，不记录请求 Body
		if lastErr := c.Errors.Last(); lastErr != nil {
			fields = append(fields, zap.Error(lastErr.Err))
		}

		switch {
		case status >= 500:
			logger.Error("http access", fields...)
		case status >= 400:
			logger.Warn("http access", fields...)
		default:
			logger.Info("http access", fields...)
		}
	}
}
