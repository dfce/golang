package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"generatego/internal/platform/datastore"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	redis *datastore.RedisClient
}

func NewRedisRepository(redis *datastore.RedisClient) *RedisRepository {
	return &RedisRepository{redis}
}

func (r *RedisRepository) Set(ctx context.Context, key, token string, expriation time.Duration) error {
	err := r.redis.Set(ctx, key, token, expriation).Err()
	return err
}

// 返回参数：
// string: 缓存值，不存在则为空“”
// bool: 代表key是否存在(true=存在, false=不存在/过期)
// error: 真正的网络故障/Redis服务异常
func (r *RedisRepository) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.redis.Get(ctx, key).Result()
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
	if !ok {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// unmarshal
	if err := json.Unmarshal([]byte(val), ptr); err != nil {
		return true, err // JSON 解析失败
	}
	return true, nil
}

// 将结构体/map 序列化为JSON写入
func (r *RedisRepository) SetObj(ctx context.Context, key string, obj any, expiration time.Duration) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, bytes, expiration).Err()
}

// 返回 json demo
func (r *RedisRepository) ScriptResJson(ctx context.Context, key string, resType any) {
	script := redis.NewScript(`
		local cjson = cjson
		return cjson.encode({
			code=0,
			balance=100
		})
	`)
	str, _ := script.Run(ctx, r.redis, []string{}).Text()

	type Resp struct {
		Code    int `json:"code"`
		Balance int `json:"balance"`
	}
	var res Resp

	json.Unmarshal([]byte(str), &res)
	fmt.Println("res:", res)
}

func (r *RedisRepository) ExecScript(ctx context.Context, s string, keys []string, val any) (string, error) {
	script := redis.NewScript(s)
	str, err := script.Run(ctx, r.redis, keys, val).Text()
	return str, err
}
