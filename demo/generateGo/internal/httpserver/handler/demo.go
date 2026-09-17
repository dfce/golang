package handler

import (
	"generatego/internal/service"
	"generatego/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DemoHandler struct {
	service *service.DemoService
	logger  *zap.Logger
}

func NewDemoHandler(service *service.DemoService, logger *zap.Logger) *DemoHandler {
	return &DemoHandler{service, logger}
}

func (d *DemoHandler) TestGet(c *gin.Context) {
	res := map[string]any{
		"msg": "demo-test",
	}
	d.logger.Info("DemoHandler TestGet", zap.Any("res", res))
	response.OK(c, res)
}

// CheckPostInfo Post测试接口
// @Summary 测试各类post请求参数数打印 以及Auth 验证
// @Description  测试各类post请求参数数打印 以及Auth 验证
// @Tags Demo模块
// @Accept json
// @Success 200 {object} response.Body
//
//	@Example 200 application/json {
//	  "code": 0,
//	  "msg": "success",
//	  "data": {
//	    "name": "Tom",
//	    "age": 18
//	  }
//	}
//
// @Router /checkpost [post]
func (h *DemoHandler) CheckPostInfo(c *gin.Context) {
	// 通过c.Request.Context() 获取被注入了超时的ctx
	// h.service.AuthInfo(c)
	h.service.AuthInfo(c.Request.Context())
	response.OK(c, "ok")
}

func (h *DemoHandler) RedisTest(c *gin.Context) {
	res, err := h.service.RedisTest(c.Request.Context())
	if err != nil {
		h.logger.Error("redis test failed", zap.Error(err))
		response.WriteError(c, err)
		return
	}
	response.OK(c, res)
}
