package user

import (
	"go-module/internal/platform/validator"
	"go-module/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Handler struct {
	service *UserSvc
}

func NewHandler(svc *UserSvc) *Handler {
	return &Handler{service: svc}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建用户账号返回用户信息
// @Tags User模块
// @Accept json
// @Produce json
// @Param body body CreateUser true "创建用户请求体：用户名 3-50 个字符且唯一，密码 8-72 个字符，邮箱可选"
// @Success 201 {object} response.Body{data=CreateUserRes} "创建成功"
// @Failure 400 {object} response.Body{data=string} "请求参数不合法"
// @Failure 409 {object} response.Body "用户名已存在"
// @Failure 500 {object} response.Body "服务器内部错误"
//
// @Router /user [post]
func (h *Handler) Create(c *gin.Context) {
	body, msg, err := validator.Validate[CreateUser](c, binding.JSON)
	if err != nil {
		response.ValidationError(c, msg)
		return
	}

	userID, err := h.service.Create(c.Request.Context(), *body)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.JSON(c, http.StatusCreated, CreateUserRes{ID: userID})
}
