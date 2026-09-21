package middleware

import (
	"errors"
	"go-module/internal/platform/datastore"
	"go-module/pkg/apperror"
	"go-module/pkg/constant"
	"go-module/pkg/jwt"
	"go-module/pkg/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

func Auth(redisSvc *datastore.RedisSvc, tokenService *jwt.Service, opts ...AuthOption) gin.HandlerFunc {
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
		if redisSvc == nil || redisSvc.Client == nil {
			c.Abort()
			response.WriteError(c, apperror.ErrRedisDisabled)
			return
		}

		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("无效的认证令牌"))
			return
		}
		token := parts[1]

		claims, err := tokenService.ParseToken(token)
		if err != nil || claims == nil {
			if err != nil {
				// AccessLog records this internal error; the client still gets a
				// generic 401 response and the raw JWT is never logged.
				_ = c.Error(err)
			}
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
			return
		}

		tokenKey := constant.AuthUserKey + strconv.FormatInt(claims.UserInfo.Id, 10)
		cacheToken, err := redisSvc.Get(c.Request.Context(), tokenKey).Result()
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
