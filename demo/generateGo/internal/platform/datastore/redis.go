package datastore

import (
	"context"
	"fmt"
	"generatego/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient = redis.Client

func OpenRedis(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
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

	logger.Info(fmt.Sprintf("redis connected adde: %s, db: %d", cfg.Addr, cfg.DB))
	return client, nil
}
