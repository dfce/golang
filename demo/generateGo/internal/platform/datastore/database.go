package datastore

import (
	"fmt"
	"generatego/internal/config"
	"generatego/internal/model"
	"generatego/internal/platform/logging"
	"generatego/pkg/util"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"

	// "gorm.io/driver/mysql"
	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Databases map[string]*gorm.DB

func OpenDatabases(configs []config.DatabaseConfig, logger *zap.Logger) (Databases, error) {
	dbs := make(Databases, len(configs))

	for _, cfg := range configs {
		dialector, err := dialectorFor(cfg)
		if err != nil {
			CloseDatabases(dbs, logger)
			return nil, err
		}

		db, err := gorm.Open(dialector, &gorm.Config{
			// Logger: gormlogger.Default.LogMode(gormlogger.Warn),
			Logger: logging.NewGormZapLogger(logger),
		})

		if err != nil {
			CloseDatabases(dbs, logger)
			return nil, fmt.Errorf("open database %q: %w", cfg.Name, err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			CloseDatabases(dbs, logger)
			return nil, fmt.Errorf("database %q sql handle: %w", cfg.Name, err)
		}

		// 连接池
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err := sqlDB.Ping(); err != nil {
			CloseDatabases(dbs, logger)
			return nil, fmt.Errorf("ping database %q: %w", cfg.Name, err)
		}

		// 本地测试：自动根据 Go 的struct 创建数据库表
		if cfg.AutoMigrate && util.IsDev(util.GetEnv("ENV", "-")) {
			_ = db.AutoMigrate(
				model.Models...,
			)
		}

		dbs[cfg.Name] = db
		logger.Info(fmt.Sprintf("database connected name: %s, driver: %s", cfg.Name, cfg.Driver))
	}

	return dbs, nil
}

func dialectorFor(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "postgres":
		return postgres.Open(cfg.DSN), nil
	// case "mysql":
	// 	return mysql.Open(cfg.DSN), nil
	// case "sqlite":
	// 	return sqlite.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver %q from %q", cfg.Driver, cfg.Name)
	}
}

func CloseDatabases(dbs Databases, logger *zap.Logger) {
	for name, db := range dbs {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Warn("get sql database failed", zap.String("name", name), zap.Error(err))
			continue
		}

		if err := sqlDB.Close(); err != nil {
			logger.Warn("close database failed", zap.String("name", name), zap.Error(err))
		}
	}
}

/*
手动出发 自动 迁移数据库
*/
func Migration(db *gorm.DB, logger *zap.Logger) {
	for _, m := range model.Models {
		if err := db.AutoMigrate(m); err != nil {
			logger.Error("表结构迁移失败", zap.Any("实体对象", m), zap.Error(err))
		}
	}
}
