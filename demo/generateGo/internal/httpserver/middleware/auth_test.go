package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"generatego/pkg/jwt"
)

func TestAuthAbortsWhenRedisIsDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService, err := jwt.NewService(jwt.Config{
		SecretKey: "test-secret-key-with-at-least-32-bytes",
		Expiry:    time.Hour,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	router := gin.New()
	handlerCalled := false
	router.GET("/protected", Auth(nil, tokenService), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if handlerCalled {
		t.Fatal("protected handler was called after authentication failure")
	}
}
