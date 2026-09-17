package healthcheck

import (
	"context"

	"generatego/internal/health"
	"generatego/internal/platform/datastore"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type databaseChecker struct {
	name     string
	db       *gorm.DB
	required bool
}

func NewDatabaseChecker(name string, db *gorm.DB, required bool) health.Checker {
	return &databaseChecker{
		name:     name,
		db:       db,
		required: required,
	}
}

func NewDatabaseCheckers(dbs datastore.Databases, required bool) []health.Checker {
	checkers := make([]health.Checker, 0, len(dbs))
	for name, db := range dbs {
		checkers = append(checkers, NewDatabaseChecker(name, db, required))
	}
	return checkers
}

func (c *databaseChecker) Name() string {
	return c.name
}

func (c *databaseChecker) Check(ctx context.Context) health.CheckResult {
	result := health.CheckResult{Required: c.required}
	if c.db == nil {
		result.Status = health.StatusFailed
		result.Error = "database is not configured"
		return result
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		result.Status = health.StatusFailed
		result.Error = err.Error()
		return result
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		result.Status = health.StatusFailed
		result.Error = err.Error()
		return result
	}

	result.Status = health.StatusOK
	return result
}

type redisChecker struct {
	client   *redis.Client
	required bool
}

func NewRedisChecker(client *redis.Client, required bool) health.Checker {
	return &redisChecker{
		client:   client,
		required: required,
	}
}

func (c *redisChecker) Name() string {
	return "redis"
}

func (c *redisChecker) Check(ctx context.Context) health.CheckResult {
	result := health.CheckResult{Required: c.required}
	if c.client == nil {
		result.Status = health.StatusDisabled
		return result
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		result.Status = health.StatusFailed
		result.Error = err.Error()
		return result
	}

	result.Status = health.StatusOK
	return result
}
