package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"generatego/internal/platform/datastore"
	"generatego/pkg/constant"
	"generatego/pkg/jwt"
	"generatego/pkg/response"
)

const AuthUserKey = "userInfo"

type AuthConfig struct {
	skip bool `comment:"没有Authorization是否跳过验证"`
}
type AuthOption func(*AuthConfig)

func WithSkip() AuthOption {
	return func(c *AuthConfig) {
		c.skip = true
	}
}

func Auth(redis *datastore.RedisClient, opts ...AuthOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := AuthConfig{}
		for _, opt := range opts {
			opt(&cfg)
		}

		header := c.GetHeader("Authorization")
		if header == "" && cfg.skip {
			c.Next()
			return
		}
		token := strings.TrimPrefix(header, constant.JwtPrefix)

		if header == "" || token == header {
			abortRes(c, -1, "Invalid Token", nil)
			return
		}

		// userInfo
		claims, err := jwt.ParseToken(token)
		if err != nil {
			abortRes(c, -1, "Invalid or Expired Token", map[string]any{})
			return
		}

		// 是否系统签发的token
		tokenKey := AuthUserKey + ":" + (strconv.FormatInt(claims.UserInfo.Id, 10))
		cacheToken, err := redis.Get(c, tokenKey).Result()
		if err != nil {
			abortRes(c, -1, "Invalid or Expired Token in system 1", map[string]any{})
			return
		}

		if cacheToken != token {
			abortRes(c, -1, "Invalid or Expired Token in system 2", "")
			return
		}

		// ctx setUserInfo
		// c.Set(AuthUserKey, claims.UserInfo)

		setCtx(c, AuthUserKey, claims.UserInfo)
		// ctx := context.WithValue(
		// 	c.Request.Context(),
		// 	AuthUserKey,
		// 	claims.UserInfo,
		// )
		// c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func abortRes(c *gin.Context, code int, msg string, data any) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}
