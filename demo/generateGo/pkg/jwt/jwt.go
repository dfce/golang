package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

/*
用户登陆授权信息
*/
type UserInfo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Claims struct {
	UserInfo
	jwtv5.RegisteredClaims
}

type Config struct {
	SecretKey string
	Expiry    time.Duration
	Issuer    string
}

type Service struct {
	secretKey []byte
	expiry    time.Duration
	issuer    string
}

func NewService(cfg Config) (*Service, error) {
	if len(cfg.SecretKey) < 32 {
		return nil, errors.New("JWT_SECRET_KEY must contain at least 32 characters")
	}
	if cfg.Expiry <= 0 {
		return nil, errors.New("JWT token expiry must be greater than zero")
	}

	return &Service{
		secretKey: []byte(cfg.SecretKey),
		expiry:    cfg.Expiry,
		issuer:    cfg.Issuer,
	}, nil
}

func (s *Service) Expiry() time.Duration {
	return s.expiry
}

func (s *Service) GenerateToken(info UserInfo) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserInfo: info,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.expiry)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			Issuer:    s.issuer,
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *Service) ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	options := []jwtv5.ParserOption{
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
	}
	if s.issuer != "" {
		options = append(options, jwtv5.WithIssuer(s.issuer))
	}

	token, err := jwtv5.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwtv5.Token) (any, error) {
			method := "<nil>"
			if t.Method != nil {
				method = t.Method.Alg()
			}
			if method != jwtv5.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected JWT signing method %q", method)
			}
			return s.secretKey, nil
		},
		options...,
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
