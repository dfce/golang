package demo

import (
	"go-module/pkg/response"

	"github.com/gin-gonic/gin"
)

type DemoHandler struct {
	service *DemoService
}

func NewDeomHandler(svc *DemoService) *DemoHandler {
	return &DemoHandler{service: svc}
}

// Ready Demo 就绪检查。
// @Summary Demo 就绪检查
// @Description 检查 Demo 模块依赖状态
// @Tags Demo模块
// @Produce json
// @Success 200 {object} response.Body{data=HealthStatusRes} "检查完成"
// @Failure 500 {object} response.Body "服务器内部错误"
// @Router /demo/ready [get]
func (h *DemoHandler) Ready(c *gin.Context) {
	res := h.service.Ready(c)
	response.OK(c, res)
}

// TestJWT JWT 测试接口。
// @Summary JWT 测试
// @Description 生成并解析 JWT，仅用于开发环境验证
// @Tags Demo模块
// @Produce json
// @Success 200 {object} response.Body{data=string} "验证成功"
// @Failure 500 {object} response.Body "服务器内部错误"
// @Router /demo/test-jwt [get]
func (h *DemoHandler) TestJWT(c *gin.Context) {
	h.service.JWTTest(c)
	response.OK(c, "OK")
}

// TestTimeout 协作式超时测试。
// @Summary 请求超时测试
// @Description 验证 Service 通过 Request Context 响应请求超时
// @Tags Demo模块
// @Produce json
// @Success 200 {object} response.Body{data=bool} "执行完成"
// @Failure 504 {object} response.Body "请求超时"
// @Router /demo/test-timeout [get]
func (h *DemoHandler) TestTimeout(c *gin.Context) {
	res, err := h.service.TimeoutTest(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.OK(c, res)
}
