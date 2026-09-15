package response

import (
	"generatego/pkg/constant"
	"net/http"

	"github.com/gin-gonic/gin"
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
	c.JSON(status, Body{
		Code:    status,
		Message: http.StatusText(status),
		Data:    data,
		TraceID: getTraceId(c),
	})
}

func ValidatErr(c *gin.Context, data string) {
	c.JSON(http.StatusBadRequest, Body{
		Code:    0,
		Message: http.StatusText(0),
		Data:    data,
		TraceID: getTraceId(c),
	})
}

func getTraceId(c *gin.Context) string {
	TraceID := c.GetString(constant.TraceName)
	if TraceID == "" {
		TraceID = c.Request.Context().Value(constant.TraceName).(string)
	}
	return TraceID
}
