package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"generatego/internal/health"

	"github.com/gin-gonic/gin"
)

type healthChecker struct {
	result health.CheckResult
}

func (c healthChecker) Name() string {
	return "dependency"
}

func (c healthChecker) Check(context.Context) health.CheckResult {
	return c.result
}

func TestHealthHandlerLivenessDoesNotDependOnReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler(health.NewService([]health.Checker{
		healthChecker{
			result: health.CheckResult{
				Status:   health.StatusFailed,
				Required: true,
			},
		},
	}, time.Second, health.NewState(false)))

	router := gin.New()
	router.GET("/livez", handler.Liveness)
	router.GET("/readyz", handler.Readiness)

	livenessRecorder := httptest.NewRecorder()
	router.ServeHTTP(livenessRecorder, httptest.NewRequest(http.MethodGet, "/livez", nil))
	if livenessRecorder.Code != http.StatusOK {
		t.Fatalf("liveness status = %d, want %d", livenessRecorder.Code, http.StatusOK)
	}

	readinessRecorder := httptest.NewRecorder()
	router.ServeHTTP(readinessRecorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if readinessRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d, want %d", readinessRecorder.Code, http.StatusServiceUnavailable)
	}
}
