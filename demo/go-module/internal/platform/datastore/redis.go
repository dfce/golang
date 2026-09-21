package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-module/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisSvc embeds the official go-redis client and adds application-specific
// helpers. All methods of *redis.Client are promoted to *RedisSvc.
type RedisSvc struct {
	*redis.Client
	logger *zap.Logger
}

// Compile-time assertions for native go-redis interfaces.
var (
	_ redis.UniversalClient = (*RedisSvc)(nil)
	_ redis.Scripter        = (*RedisSvc)(nil)
)

func OpenRedis(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisSvc, error) {
	return NewRedisSvc(ctx, cfg, logger)
}

func NewRedisSvc(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisSvc, error) {
	if !cfg.Enabled {
		logger.Info("redis disabled")
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	logger.Info(fmt.Sprintf("redis connected addr: %s, db: %d", cfg.Addr, cfg.DB))
	return &RedisSvc{
		Client: client,
		logger: logger,
	}, nil
}

// 返回参数：
// string: 缓存值，不存在则为空“”
// bool: 代表key是否存在(true=存在, false=不存在/过期)
// error: 真正的网络故障/Redis服务异常
// GetString returns the cached string value, whether it exists, and an
// infrastructure error. It is intentionally not named Get so the native
// redis.Client.Get method remains promoted and available.
func (r *RedisSvc) GetString(ctx context.Context, key string) (string, bool, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// 读取JSON缓存并反序列化到结构体指针
// ptr 必须传一个go结构体指针如：&user
func (r *RedisSvc) GetObj(ctx context.Context, key string, ptr any) (bool, error) {
	val, ok, err := r.GetString(ctx, key)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	if err := json.Unmarshal([]byte(val), ptr); err != nil {
		return true, err
	}
	return true, nil
}

// 将结构体/map 序列化为JSON写入
func (r *RedisSvc) SetObj(ctx context.Context, key string, obj any, expiration time.Duration) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = r.Client.Set(ctx, key, payload, expiration).Err()
	if err == nil && r.logger != nil {
		r.logger.Debug("redis JSON value stored", zap.String("key", key))
	}
	return err
}

func (r *RedisSvc) ExecScript(ctx context.Context, script string, keys []string, values ...any) (string, error) {
	return redis.NewScript(script).Run(ctx, r, keys, values...).Text()
}
