package repository

import (
	"context"
	"encoding/json"
	"time"

	"generatego/internal/platform/datastore"
	"generatego/pkg/apperror"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	redis *datastore.RedisClient
}

func NewRedisRepository(redis *datastore.RedisClient) *RedisRepository {
	return &RedisRepository{redis: redis}
}

func (r *RedisRepository) Enabled() bool {
	return r != nil && r.redis != nil
}

func (r *RedisRepository) client() (*datastore.RedisClient, error) {
	if r == nil || r.redis == nil {
		return nil, apperror.ErrRedisDisabled
	}
	return r.redis, nil
}

func (r *RedisRepository) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	client, err := r.client()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// 返回参数：
// string: 缓存值，不存在则为空“”
// bool: 代表key是否存在(true=存在, false=不存在/过期)
// error: 真正的网络故障/Redis服务异常
// Get returns the cached value, whether it exists, and an infrastructure error.
func (r *RedisRepository) Get(ctx context.Context, key string) (string, bool, error) {
	client, err := r.client()
	if err != nil {
		return "", false, err
	}

	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// 读取JSON缓存并反序列化到结构体指针
// ptr 必须传一个go结构体指针如：&user
func (r *RedisRepository) GetObj(ctx context.Context, key string, ptr any) (bool, error) {
	val, ok, err := r.Get(ctx, key)
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
func (r *RedisRepository) SetObj(ctx context.Context, key string, obj any, expiration time.Duration) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, string(bytes), expiration)
}

func (r *RedisRepository) ExecScript(ctx context.Context, script string, keys []string, values ...any) (string, error) {
	client, err := r.client()
	if err != nil {
		return "", err
	}

	return redis.NewScript(script).Run(ctx, client, keys, values...).Text()
}
