package tokenservice

import (
	"context"
	"errors"
	"time"

	"generatego/internal/port"
	"generatego/pkg/jwt"
)

type JWTService struct {
	service *jwt.Service
}

func NewJWTService(service *jwt.Service) *JWTService {
	if service == nil {
		return nil
	}
	return &JWTService{service: service}
}

func (s *JWTService) Issue(_ context.Context, principal port.Principal) (string, time.Duration, error) {
	if s == nil || s.service == nil {
		return "", 0, errors.New("JWT service is not configured")
	}
	token, err := s.service.GenerateToken(jwt.UserInfo{
		Id:   principal.UserID,
		Name: principal.Username,
	})
	return token, s.service.Expiry(), err
}

func (s *JWTService) Verify(_ context.Context, token string) (port.Principal, error) {
	if s == nil || s.service == nil {
		return port.Principal{}, errors.New("JWT service is not configured")
	}
	claims, err := s.service.ParseToken(token)
	if err != nil {
		return port.Principal{}, err
	}
	return port.Principal{
		UserID:   claims.UserInfo.Id,
		Username: claims.UserInfo.Name,
	}, nil
}

var _ port.TokenService = (*JWTService)(nil)
