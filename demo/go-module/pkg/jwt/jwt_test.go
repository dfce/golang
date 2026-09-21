package jwt

import (
	"strings"
	"testing"
	"time"
)

func testService(t *testing.T) *Service {
	t.Helper()

	service, err := NewService(Config{
		SecretKey: "test-secret-key-that-is-at-least-32-bytes",
		Expiry:    time.Hour,
		Issuer:    "test",
	})
	if err != nil {
		t.Fatalf("create JWT service: %v", err)
	}
	return service
}

func TestGenerateAndParseToken(t *testing.T) {
	service := testService(t)

	token, err := service.GenerateToken(UserInfo{
		Id:   36,
		Name: "Alice",
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("parse generated token: %v", err)
	}
	if claims == nil {
		t.Fatal("claims must not be nil for a valid token")
	}
	if claims.UserInfo.Id != 36 || claims.UserInfo.Name != "Alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestParseTokenReturnsCauseForInvalidToken(t *testing.T) {
	service := testService(t)

	claims, err := service.ParseToken("not-a-jwt")
	if err == nil {
		t.Fatal("expected invalid token error")
	}
	if claims != nil {
		t.Fatalf("expected nil claims, got %+v", claims)
	}
	if !strings.Contains(err.Error(), "parse JWT token") {
		t.Fatalf("expected parse context in error, got %v", err)
	}
}
