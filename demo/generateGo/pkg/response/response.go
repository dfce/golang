package response

import (
	"context"
	"errors"
	"net/http"

	"generatego/pkg/apperror"

	"github.com/gin-gonic/gin"

	"generatego/pkg/util"
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}

func JSON(c *gin.Context, status int, data any) {
	write(c, status, http.StatusText(status), data)
}

func ValidationError(c *gin.Context, message string) {
	write(c, http.StatusBadRequest, "请求参数校验失败", message)
}

// WriteError converts internal errors into a consistent, client-safe response.
// The original error should be logged by the caller when necessary.
func WriteError(c *gin.Context, err error) {
	if err == nil {
		InternalError(c)
		return
	}

	status := http.StatusInternalServerError
	message := "服务器内部错误"

	var appErr *apperror.Error
	switch {
	case errors.Is(err, apperror.ErrRedisDisabled):
		status = http.StatusServiceUnavailable
		message = "Redis 未启用，当前接口不可用"
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
		message = "请求处理超时"
	case errors.As(err, &appErr):
		status = appErr.Status
		message = appErr.Message
		if status >= http.StatusInternalServerError && status != http.StatusServiceUnavailable {
			message = "服务器内部错误"
		}
	}

	write(c, status, message, nil)
}

func InternalError(c *gin.Context) {
	write(c, http.StatusInternalServerError, "服务器内部错误", nil)
}

func write(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Body{
		Code:    status,
		Message: message,
		Data:    data,
		TraceID: util.TraceID(c),
	})
}
