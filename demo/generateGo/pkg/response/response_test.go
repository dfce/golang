package response

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"generatego/pkg/apperror"
	"generatego/pkg/constant"
)

func TestWriteErrorHidesInternalErrorAndIncludesTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(context.WithValue(
		request.Context(),
		constant.TraceName,
		"trace-123",
	))

	context, _ := gin.CreateTestContext(recorder)
	context.Request = request
	WriteError(context, apperror.Wrap(errors.New("database password leaked"), http.StatusInternalServerError, "database failed"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	var body Body
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Message != "服务器内部错误" {
		t.Fatalf("message = %q, want internal error message", body.Message)
	}
	if body.TraceID != "trace-123" {
		t.Fatalf("trace id = %q, want %q", body.TraceID, "trace-123")
	}
	if body.Data != nil {
		t.Fatalf("data = %#v, want nil", body.Data)
	}
	if string(recorder.Body.Bytes()) == "" {
		t.Fatal("response body is empty")
	}
}

func TestWriteErrorMapsDisabledRedisTo503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(context, apperror.ErrRedisDisabled)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
