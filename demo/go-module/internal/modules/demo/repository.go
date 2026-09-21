package demo

import (
	"context"
	"go-module/internal/platform/datastore"

	"gorm.io/gorm"
)

type DemoRepository struct {
	db    *gorm.DB
	redis *datastore.RedisSvc
}

func NewDemoRepository(db *gorm.DB, redis *datastore.RedisSvc) *DemoRepository {
	return &DemoRepository{db, redis}
}

func (d *DemoRepository) Ready(ctx context.Context) {

}

func (d *DemoRepository) CheckRedis(ctx context.Context) CheckResult {
	if d.redis == nil {
		return CheckResult{Enabled: false, OK: true}
	}

	if err := d.redis.Ping(ctx).Err(); err != nil {
		return CheckResult{Enabled: true, OK: false, Error: err.Error()}
	}
	return CheckResult{Enabled: true, OK: true}
}
