package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 配置全局超时中间件

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 基于当前 Request Context 创建一个带有超时的 context
		// cancel 必须被调用，否则会导致底层定时器内存泄漏
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 2. 将带有超时现在的新ctx 重新注入会Gin的Request
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 超时触发
		if ctx.Err() == context.DeadlineExceeded {
			// 如果此时Header还没响应前端，输出 504
			if !c.Writer.Written() {
				c.AbortWithStatusJSON(
					http.StatusGatewayTimeout,
					gin.H{
						"code":    http.StatusGatewayTimeout,
						"message": "接口请求超时， 服务中断",
					})
			}
			return
		}
	}
}

func setCtx(c *gin.Context, key string, val any) {
	ctx := context.WithValue(
		c.Request.Context(),
		key, val,
	)
	c.Request = c.Request.WithContext(ctx)
}
