package payment

import (
	"go-module/internal/platform/validator"
	"go-module/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Handler struct {
	service *PaymentSvc
}

func NewHandler(svc *PaymentSvc) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Deposit(c *gin.Context) {
	body, msg, err := validator.Validate[PayReq](c, binding.JSON)
	if err != nil {
		response.ValidationError(c, msg)
		return
	}
	h.service.Pay(c, *body)
}
