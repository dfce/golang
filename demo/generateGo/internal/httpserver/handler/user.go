package handler

import (
	"generatego/internal/model"
	"generatego/internal/service"
	"generatego/pkg/response"
	myvalidator "generatego/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{service, logger}
}

func (u *UserHandler) Create(c *gin.Context) {

	body, msg, err := myvalidator.ValidatParam[model.CreateUser](c, binding.JSON)
	if err != nil {
		response.ValidationError(c, msg)
		return
	}

	if _, err := u.service.Create(c.Request.Context(), *body); err != nil {
		u.logger.Error("create user failed", zap.Error(err))
		response.WriteError(c, err)
		return
	}
	response.OK(c, "")
}

func (u *UserHandler) Login(c *gin.Context) {

	body, msg, err := myvalidator.ValidatParam[model.UserLogin](c, binding.JSON)
	if err != nil {
		response.ValidationError(c, msg)
		return
	}

	token, err := u.service.Login(c.Request.Context(), *body)
	if err != nil {
		u.logger.Error("user login failed", zap.Error(err))
		response.WriteError(c, err)
		return
	}
	response.OK(c, token)
}

func (h *UserHandler) List(c *gin.Context) {

	req, msg, err := myvalidator.ValidatParam[model.GetUser](c, binding.Query)
	if err != nil {
		response.ValidationError(c, msg)
		return
	}

	// 通过c.Request.Context() 获取被注入了超时的ctx
	list, err := h.service.List(c.Request.Context(), *req)
	if err != nil {
		h.logger.Error("list users failed", zap.Error(err))
		response.WriteError(c, err)
		return
	}

	response.OK(c, list)
}
