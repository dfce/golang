package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"generatego/internal/repository"
	"generatego/pkg/apperror"

	"go.uber.org/zap"
)

type HealthStatus struct {
	Status    string                            `json:"status"`
	Databases map[string]repository.CheckResult `json:"databases"`
	Redis     repository.CheckResult            `json:"redis"`
}

type RedisTestResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Stock int    `json:"stock"`
}

type DemoService struct {
	*BaseService
	repo  *repository.DemoRepository
	redis *repository.RedisRepository
}

func NewDemoService(repo *repository.DemoRepository, redis *repository.RedisRepository, logger *zap.Logger) *DemoService {
	return &DemoService{
		BaseService: &BaseService{Logger: logger},
		repo:        repo,
		redis:       redis,
	}
}

func (d *DemoService) Ready(ctx context.Context) HealthStatus {
	readyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	d.CtxLog(ctx).Info("checking application dependencies")

	status := HealthStatus{
		Status:    "ok",
		Databases: d.repo.CheckDatabases(readyCtx),
		Redis:     d.repo.CheckRedis(readyCtx),
	}

	for _, result := range status.Databases {
		if !result.OK {
			status.Status = "degraded"
			return status
		}
	}

	if !status.Redis.OK {
		status.Status = "degraded"
	}
	return status
}

func (d *DemoService) AuthInfo(ctx context.Context) {
	userInfo, ok := d.Userinfo(ctx)
	if !ok {
		d.CtxLog(ctx).Warn("authenticated user information is missing")
		return
	}
	d.CtxLog(ctx).Debug("current user", zap.Int64("id", userInfo.Id))
}

func (s *DemoService) RedisTest(ctx context.Context) (RedisTestResponse, error) {
	script := `
		local decrby = tonumber(ARGV[1])
		local initStock = tonumber(ARGV[2])
		local expire = tonumber(ARGV[3])
		redis.call("SET", KEYS[1], initStock, "EX", expire, "NX")

		local stock = tonumber(redis.call("GET", KEYS[1]))
		if stock < decrby then
			return cjson.encode({code=-1,msg="库存不足",stock=stock})
		end

		if stock <= 0 then
			return cjson.encode({code=-1,msg="今日机会已用完",stock=stock})
		end

		local remain = redis.call("DECRBY", KEYS[1], decrby)
		return cjson.encode({code=0,msg="使用成功",stock=remain})
	`

	res, err := s.redis.ExecScript(ctx, script, []string{"test:deduct:1"}, 3, 10, 300)
	if err != nil {
		if errors.Is(err, apperror.ErrRedisDisabled) {
			return RedisTestResponse{}, err
		}
		return RedisTestResponse{}, apperror.Wrap(err, http.StatusServiceUnavailable, "Redis 服务暂不可用")
	}

	var response RedisTestResponse
	if err := json.Unmarshal([]byte(res), &response); err != nil {
		return RedisTestResponse{}, apperror.Wrap(err, http.StatusInternalServerError, "Redis 返回数据格式错误")
	}
	return response, nil
}
