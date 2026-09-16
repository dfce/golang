package middleware

import (
	"generatego/pkg/constant"
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

		setCtx(c, constant.TraceName, traceID)
		c.Writer.Header().Set(constant.TraceHeader, traceID)
		c.Next()
	}
}
