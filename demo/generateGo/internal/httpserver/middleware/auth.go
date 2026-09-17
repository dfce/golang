package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"generatego/internal/port"
	"generatego/pkg/apperror"
	"generatego/pkg/constant"
	"generatego/pkg/response"
)

type AuthConfig struct {
	enabled bool
	skip    bool `comment:"没有Authorization是否跳过验证"`
}
type AuthOption func(*AuthConfig)

func WithEnabled(enabled bool) AuthOption {
	return func(c *AuthConfig) {
		c.enabled = enabled
	}
}

func WithSkip() AuthOption {
	return func(c *AuthConfig) {
		c.skip = true
	}
}

func Auth(sessions port.SessionStore, tokenService port.TokenService, opts ...AuthOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := AuthConfig{enabled: true}
		for _, opt := range opts {
			opt(&cfg)
		}

		if !cfg.enabled {
			c.Next()
			return
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
		if sessions == nil {
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

		principal, err := tokenService.Verify(c.Request.Context(), token)
		if err != nil {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
			return
		}

		tokenKey := constant.AuthUserKey + ":" + strconv.FormatInt(principal.UserID, 10)
		cacheToken, found, err := sessions.Get(c.Request.Context(), tokenKey)
		if err != nil {
			c.Abort()
			response.WriteError(c, apperror.Wrap(err, http.StatusServiceUnavailable, "认证服务暂不可用"))
			return
		}

		if !found || cacheToken != token {
			c.Abort()
			response.WriteError(c, apperror.Unauthorized("认证令牌无效或已过期"))
			return
		}

		setCtx(c, constant.AuthUserKey, principal)
		c.Next()
	}
}
