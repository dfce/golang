package demo

import (
	"context"
	"go-module/internal/platform/datastore"
	"go-module/internal/platform/logging"
	"go-module/pkg/base"
	"go-module/pkg/constant"
	"go-module/pkg/jwt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DemoService struct {
	*base.BaseService
	repo         *DemoRepository
	logger       *zap.Logger
	tokenService *jwt.Service
}

func NewDemoService(repo *DemoRepository, redis *datastore.RedisClient, logger *zap.Logger, tokenService *jwt.Service) *DemoService {
	return &DemoService{
		BaseService:  &base.BaseService{Logger: logger},
		repo:         repo,
		logger:       logger,
		tokenService: tokenService,
	}
}

func (d *DemoService) Ready(ctx *gin.Context) HealthStatusRes {
	// res :=
	dCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	d.CtxLog(ctx).Info("checking application dependencies")

	status := HealthStatusRes{
		Status: "ok",
		Redis:  d.repo.CheckRedis(dCtx),
	}

	return status
}

func (d *DemoService) JWTTest(ctx *gin.Context) {
	// dCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	d.CtxLog(ctx).Info("checking JWTTest")

	// 生产认证令牌
	const userID int64 = 36
	token, err := d.tokenService.GenerateToken(jwt.UserInfo{
		Id:   userID,
		Name: "Alice",
	})
	if err != nil {
		d.CtxLog(ctx).Info("生产认证令牌失败")
		// return "", apperror.Wrap(err, http.StatusInternalServerError, "生产Token失败")
	}

	// 写入Redis
	tokenKey := constant.AuthUserKey + strconv.FormatInt(userID, 10)
	expireDuration := d.tokenService.Expiry()
	d.CtxLog(ctx).Info("写入Redis", logging.KV2Fields("tokenKey", tokenKey, "expire", expireDuration, "token", token)...)
	// if err := redis.Set(dCtx, tokenKey, token, expireDuration); err != nil {
	// 	d.CtxLog(ctx).Info("写入Redis失败")
	// 	// return "", apperror.Wrap(err, http.StatusInternalServerError, "写入Redis失败")
	// }

	// 验证 解析 token
	claims, err := d.tokenService.ParseToken(token)
	if err != nil {
		d.CtxLog(ctx).Info("认证令牌无效或已过期")
		// return "", apperror.Wrap(err, http.StatusInternalServerError, "认证令牌无效或已过期")
	}
	d.CtxLog(ctx).Info("认证令牌解析完成", logging.KV2Fields("claims", claims)...)
}

func (d *DemoService) TimeoutTest(ctx context.Context) (bool, error) {

	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 1, "tokenKey", time.Now())...)
	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 2, "tokenKey", time.Now())...)
	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 3, "tokenKey", time.Now())...)
	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 4, "tokenKey", time.Now())...)

	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 5, "tokenKey", time.Now())...)
	// time.Sleep(510 * time.Millisecond)
	// d.CtxLog(ctx).Info("TimeoutTest", logging.KV2Fields("index", 6, "tokenKey", time.Now())...)

	for i := 0; i < 6; i++ {
		if err := waitContext(ctx, 510*time.Millisecond); err != nil {
			return false, err
		}

		d.CtxLog(ctx).Info("TimeoutTest")
	}
	return true, nil
}

func waitContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		// 正常等待
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
