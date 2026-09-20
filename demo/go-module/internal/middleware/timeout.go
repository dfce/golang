package middleware

import (
	"context"
	"errors"
	"go-module/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// 配置全局超时中间件

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 {
			c.Next()
			return
		}
		// 1. 基于当前 Request Context 创建一个带有超时的 context
		// cancel 必须被调用，否则会导致底层定时器内存泄漏
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 2. 将带有超时现在的新ctx 重新注入回Gin的Request
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 仅作为兜底： Handler 返回后仍未写响应时返回 504
		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			response.WriteError(c, ctx.Err())
		}
	}
}
