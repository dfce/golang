package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"generatego/pkg/constant"
)

func TestTraceIDUsesTraceHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(constant.TraceHeader, "trace-123")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get(constant.TraceHeader) != "trace-123" {
		t.Fatalf("trace header = %q, want %q", recorder.Header().Get(constant.TraceHeader), "trace-123")
	}
	if recorder.Header().Get(constant.TraceName) != "" {
		t.Fatalf("unexpected internal trace header %q", recorder.Header().Get(constant.TraceName))
	}
}
