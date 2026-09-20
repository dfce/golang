package modules

import (
	"encoding/json"
	"go-module/internal/config"
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "go-module/swdocs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestSwaggerDocJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App: config.AppConfig{
			Name: "test",
			ENV:  "dev",
		},
		HTTP: config.HTTPConfig{
			RequestTimeout: time.Second,
		},
		Log: config.LogConfig{
			Level: "error",
			Mode:  "2",
		},
	}

	tokenService, err := jwt.NewService(jwt.Config{
		SecretKey: "test-secret-key-that-is-at-least-32-bytes",
		Expiry:    time.Hour,
		Issuer:    "test",
	})
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}

	router := BuildAppEngine(
		cfg,
		datastore.Databases{"primary": nil},
		nil,
		zap.NewNop(),
		tokenService,
	)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var doc struct {
		BasePath    string                 `json:"basePath"`
		Paths       map[string]interface{} `json:"paths"`
		Definitions map[string]interface{} `json:"definitions"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode swagger document: %v", err)
	}
	if doc.BasePath != "/api/v1" {
		t.Fatalf("expected basePath /api/v1, got %q", doc.BasePath)
	}
	if _, ok := doc.Paths["/user"]; !ok {
		t.Fatalf("expected /user in swagger paths")
	}
	if _, ok := doc.Paths["/demo/ready"]; !ok {
		t.Fatalf("expected /demo/ready in swagger paths")
	}
	for _, name := range []string{
		"response.Body",
		"demo.HealthStatusRes",
		"demo.CheckResult",
		"user.CreateUser",
		"user.CreateUserRes",
	} {
		if _, ok := doc.Definitions[name]; !ok {
			t.Fatalf("expected %s in swagger definitions", name)
		}
	}
}
