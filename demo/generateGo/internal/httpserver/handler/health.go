package handler

import (
	"net/http"

	"generatego/internal/health"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	readiness *health.Service
}

func NewHealthHandler(readiness *health.Service) *HealthHandler {
	return &HealthHandler{readiness: readiness}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	if h.readiness == nil {
		c.JSON(http.StatusServiceUnavailable, health.Report{
			Status: health.StatusNotReady,
			Checks: map[string]health.CheckResult{},
		})
		return
	}

	report := h.readiness.Check(c.Request.Context())
	status := http.StatusOK
	if report.Status == health.StatusNotReady {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, report)
}
