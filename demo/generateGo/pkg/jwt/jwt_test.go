package jwt_test

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"generatego/pkg/jwt"
)

const testSecret = "test-secret-key-with-at-least-32-bytes"

func TestServiceRoundTrip(t *testing.T) {
	service, err := jwt.NewService(jwt.Config{
		SecretKey: testSecret,
		Expiry:    time.Hour,
		Issuer:    "test",
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	const userID int64 = 42
	token, err := service.GenerateToken(jwt.UserInfo{Id: userID, Name: "alice"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserInfo.Id != userID {
		t.Fatalf("claims user id = %d, want %d", claims.UserInfo.Id, userID)
	}
	if claims.Issuer != "test" {
		t.Fatalf("claims issuer = %q, want %q", claims.Issuer, "test")
	}
}

func TestServiceRejectsNonHS256(t *testing.T) {
	service, err := jwt.NewService(jwt.Config{
		SecretKey: testSecret,
		Expiry:    time.Hour,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS512, &jwt.Claims{
		UserInfo: jwt.UserInfo{Id: 42, Name: "alice"},
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := service.ParseToken(signed); err == nil {
		t.Fatal("ParseToken() accepted HS512 token")
	}
}
