package repository

import (
	"context"
	"generatego/internal/model"
	"generatego/internal/platform/datastore"

	"gorm.io/gorm"
)

type DemoRepository struct {
	dbs   datastore.Databases
	redis *datastore.RedisClient
}

type CheckResult struct {
	Enabled bool   `json:"enabled"`
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

func NewDemoRepository(dbs datastore.Databases, redis *datastore.RedisClient) *DemoRepository {
	return &DemoRepository{dbs, redis}
	// return &DemoRepository{dbs:dbs, redis:redis}
}

func (d *DemoRepository) CheckDatabases(ctx context.Context) map[string]CheckResult {
	results := make(map[string]CheckResult, len(d.dbs))

	for name, db := range d.dbs {
		sqlDB, err := db.DB()
		if err != nil {
			results[name] = CheckResult{OK: false, Error: err.Error()}
			continue
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			results[name] = CheckResult{OK: false, Error: err.Error()}
			continue
		}
		results[name] = CheckResult{OK: true}
	}
	return results
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

func (r *DemoRepository) TestTransaction() error {
	return r.dbs["primary"].Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.User{}).Error; err != nil {
			return err
		}

		// tx.Model(&model.User{}).Where("id=?", 1).Update("email", "new_email")
		// // 计数器 用UpdateColumn
		// tx.Model(&model.User{}).Where("id=?", 1).UpdateColumn("likes", gorm.Expr("likes + ?", 1))
		// tx.Model(&model.User{}).Where("id=? and blance >= 100", 1).UpdateColumn("blance", gorm.Expr("blance - ?", 100))
		// tx.Model(&model.User{}).Where("id=?", 1).UpdateColumns(map[string]any{
		// 	"status": 1,
		// })

		// result := tx.Model(&model.User{}).Where("id=? and blance >= 100", 1).UpdateColumn("blance", gorm.Expr("blance - ?", 100))
		// if result.Error != nil {
		// 	return result.Error
		// }
		// if result.RowsAffected != 1 {
		// 	return errors.New("余额不足")
		// }
		return nil
	})
}
