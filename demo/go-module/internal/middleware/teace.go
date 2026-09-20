package middleware

import (
	"context"
	"go-module/pkg/constant"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := strings.TrimSpace(c.GetHeader(constant.TraceHeader))
		if traceID == "" {
			traceID = strings.ReplaceAll(uuid.NewString(), "-", "") // uuid.NewString()
		}

		setCtx(c, constant.TraceKey, traceID)
		c.Writer.Header().Set(constant.TraceHeader, traceID)
		c.Next()
	}
}

func setCtx(c *gin.Context, key string, val any) {
	ctx := context.WithValue(
		c.Request.Context(),
		key, val,
	)
	c.Request = c.Request.WithContext(ctx)
}
