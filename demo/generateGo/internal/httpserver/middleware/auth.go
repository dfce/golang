package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"generatego/pkg/apperror"
	"generatego/pkg/constant"
	"generatego/pkg/jwt"
	"generatego/pkg/response"
)

type AuthConfig struct {
	skip bool `comment:"没有Authorization是否跳过验证"`
}
type AuthOption func(*AuthConfig)

func WithSkip() AuthOption {
	return func(c *AuthConfig) {
		c.skip = true
	}
}

func Auth(redisClient *redis.Client, tokenService *jwt.Service, opts ...AuthOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := AuthConfig{}
		for _, opt := range opts {
			opt(&cfg)
		}

		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" && cfg.skip {
			c.Next()
			return
		}

		if tokenService == nil {
			c.Abort()
			response.WriteError(c, apperror.ServiceUnavailable("JWT 认证服务未配置"))
			return
		}
		if redisClient == nil {
			c.Abort()
			response.WriteError(c, apperror.ErrRedisDisabled)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, constant.JwtPrefix))
		if header == "" || token == header {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("无效的认证令牌"))
			return
		}

		claims, err := tokenService.ParseToken(token)
		if err != nil {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
			return
		}

		tokenKey := constant.AuthUserKey + ":" + strconv.FormatInt(claims.UserInfo.Id, 10)
		cacheToken, err := redisClient.Get(c.Request.Context(), tokenKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				c.Abort()
				response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
				return
			}
			c.Abort()
			response.WriteError(c, apperror.Wrap(err, http.StatusServiceUnavailable, "认证服务暂不可用"))
			return
		}

		if cacheToken != token {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
			return
		}

		setCtx(c, constant.AuthUserKey, claims.UserInfo)
		c.Next()
	}
}
