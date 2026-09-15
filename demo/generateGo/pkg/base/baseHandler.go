package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"generatego/pkg/constant"
)

type BaseHandler struct {
	// Logger *zap.Logger
}

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// func OK(c *gin.Context, data any) {
func (b *BaseHandler) OK(c *gin.Context, data any) {
	b.JSON(c, http.StatusOK, data)
}

func (b *BaseHandler) JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Body{
		Code:    status,
		Message: http.StatusText(status),
		Data:    data,
		TraceID: c.GetString(constant.TraceName),
	})
}

func (b *BaseHandler) ValidatErr(c *gin.Context, data string) {

	c.JSON(http.StatusBadRequest, Body{
		Code:    0,
		Message: http.StatusText(0),
		Data:    data,
		TraceID: c.GetString(constant.TraceName),
	})
}
