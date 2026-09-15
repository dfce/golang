package service

import (
	"context"
	"encoding/json"
	"fmt"
	"generatego/internal/repository"
	"time"

	"go.uber.org/zap"
)

type HealthStatus struct {
	Status    string                            `json:"status"`
	Databases map[string]repository.CheckResult `json:"databases"`
	Redis     repository.CheckResult            `json:"redis"`
}

type DemoService struct {
	*BaseService
	repo  *repository.DemoRepository
	redis *repository.RedisRepository
}

func NewDemoService(repo *repository.DemoRepository, redis *repository.RedisRepository, logger *zap.Logger) *DemoService {
	return &DemoService{&BaseService{Logger: logger}, repo, redis}
}

func (d *DemoService) Ready(ctx context.Context) HealthStatus {
	readyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	d.CtxLog(ctx).Info("测试 带入每次请求的 traceID")

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
		return status
	}
	return status
}

func (d *DemoService) AuthInfo(ctx context.Context) {
	userinfo := d.Userinfo(ctx)
	fmt.Printf("cur user: %+v, Id: %d \n", userinfo, userinfo.Id)
}

func (s *DemoService) RedisTest(ctx context.Context) any {
	// 模拟 每日 扣减用户数据, 当天未使用则初始化后再扣减
	// key= test:deduct:${uid}
	// val = ["扣除值","初始化值","key过期时间"]
	//  更优的 script 逻辑, 一次执行 setnx
	script := `
		local cjson = cjson

		local decrby = tonumber(ARGV[1])
		local initStock = tonumber(ARGV[2])
		local expire = tonumber(ARGV[3])
		redis.call("SET", KEYS[1], initStock, "EX", expire, "NX")

		local stock = tonumber(redis.call("GET", KEYS[1]))
		
		-- 防止 decrby > 1 时， 扣为负数
		if stock < decrby then
			return cjson.encode({code=-1,msg="库存不足",stock=stock})
		end

		if stock <= 0 then
			return cjson.encode({code=-1,msg="今日机会已用完",stock=stock})
		end

		-- 扣除使用
		local remain = redis.call("DECRBY", KEYS[1], decrby)
		return cjson.encode({code=0,msg="使用成功",stock=remain})
	`

	keys := []string{"test:deduct:1"}
	vals := []any{3, 10, 300}
	res, err := s.redis.ExecScript(ctx, script, keys, vals)
	if err != nil {
		fmt.Println("DemoService RedisTest error", err)
	}
	type Resp struct {
		// code=-1,msg="今日机会已用完",stock=stock
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Stock int    `json:"stock"`
	}

	var resp Resp
	if err := json.Unmarshal([]byte(res), &resp); err != nil {
		fmt.Println("DemoService RedisTest Unmarshal error", err)
		return nil
	}

	return resp
}
