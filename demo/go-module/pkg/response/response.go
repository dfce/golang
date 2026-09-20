package response

import (
	"context"
	"errors"
	"go-module/pkg/apperror"
	"go-module/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	// HTTP 状态码。
	Code int `json:"code" example:"200"`
	// 面向客户端的提示信息。
	Message string `json:"message" example:"OK"`
	// 业务数据。具体结构由接口响应注解中的 data 类型决定。
	Data any `json:"data,omitempty"`
	// 请求链路 ID。
	TraceID string `json:"trace_id,omitempty" example:"a1b2c3d4"`
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}
func JSON(c *gin.Context, status int, data any) {
	write(c, status, http.StatusText(status), data)
}

func InternalError(c *gin.Context) {
	write(c, http.StatusInternalServerError, "服务器内部错误", nil)
}

func ValidationError(c *gin.Context, message string) {
	write(c, http.StatusBadRequest, "请求参数校验错误", message)
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
		message = "Redis 未启用， 当前服务不可用"
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
		message = "请求超时"
	case errors.As(err, &appErr):
		status = appErr.Status
		message = appErr.Message

		if status >= http.StatusInternalServerError && status != http.StatusServiceUnavailable {
			message = "服务器内部错误"
		}
	}
	write(c, status, message, nil)
}

func write(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Body{
		Code:    status,
		Message: message,
		Data:    data,
		TraceID: util.TraceID(c),
	})
}
