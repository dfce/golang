package jwt

import (
	"errors"
	"generatego/pkg/constant"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
用户登陆授权信息
*/
type UserInfo struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	// Level int    `json:"level"` // 可能变更不及时
	// ...
}

type Claims struct {
	UserInfo
	jwt.RegisteredClaims
}

func GenerateToken(info UserInfo) (string, error) {
	expirationTime := time.Now().Add(time.Second * time.Duration(constant.JwtExpire))

	claims := &Claims{
		UserInfo: info,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(constant.JwtKey)
	if err != nil {
		return "", err
	}

	return str, nil
	// return constant.JwtPrefix + str, nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return constant.JwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}
