package config

import (
	"errors"
	"fmt"
	"generatego/pkg/util"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

/*
	configs 配置中心
*/

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Log       LogConfig
	Auth      AuthConfig
	Databases []DatabaseConfig
	Redis     RedisConfig
}

type AppConfig struct {
	Name string
	ENV  string
}
type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func (c HTTPConfig) Addr() string {
	return ":" + c.Port
}

type LogConfig struct {
	Level string
}

type AuthConfig struct {
	Enabled   bool
	Token     string
	SecretKey string
}

type DatabaseConfig struct {
	AutoMigrate     bool
	Name            string
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Enabled      bool
	Addr         string
	Username     string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() (*Config, error) {
	env := util.GetEnv("ENV", "dev")
	// 开发环境获取本地 .env 的配置项 加载Config
	if util.IsDev(env) {
		_ = godotenv.Load()
		env = util.GetEnv("ENV", env)
	}

	cfg := &Config{
		App: AppConfig{
			Name: util.GetEnv("APP_NAME", "generate_go"),
			ENV:  env,
		},
		HTTP: HTTPConfig{
			Port:         util.GetEnv("APP_PORT", "8080"),
			ReadTimeout:  util.DurationEnv("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: util.DurationEnv("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  util.DurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Log: LogConfig{
			Level: util.GetEnv("LOG_LEVEL", "info"),
		},
		Redis: RedisConfig{
			Enabled:      util.BoolEnv("REDIS_ENABLED", false),
			Addr:         util.GetEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Username:     util.GetEnv("REDIS_USERNAME", ""),
			Password:     util.GetEnv("REDIS_PASSWORD", ""),
			DB:           util.IntEnv("REDIS_DB", 0),
			DialTimeout:  util.DurationEnv("REDIS_DIAL_TIMEOUT", 3*time.Second),
			WriteTimeout: util.DurationEnv("REDIS_WRITE_TIMEOUT", 2*time.Second),
			ReadTimeout:  util.DurationEnv("REDIS_READ_TIMEOUT", 2*time.Second),
		},
	}

	// databases
	databases, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}
	cfg.Databases = databases

	return cfg, nil
}

func loadDatabaseConfig() ([]DatabaseConfig, error) {
	names := util.SplitSCV(util.GetEnv("DB_CONNECTIONS", "primary"))
	if len(names) == 0 {
		return nil, errors.New("DB_CONNECTIONS must contain at least one database name")
	}

	// 多数据连接配置
	databases := make([]DatabaseConfig, 0, len(names))
	for _, name := range names {
		prefix := "DB_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_"
		driver := util.GetEnv(prefix+"DRIVER", defaultDBDriver(name))
		dsn := util.GetEnv(prefix+"DSN", defaultDBDSN(name, driver))

		if driver == "" || dsn == "" {
			return nil, fmt.Errorf("database %q requires %sDRIVER and %sDSN", name, prefix, prefix)
		}

		databases = append(databases, DatabaseConfig{
			Name:            name,
			Driver:          strings.ToLower(driver),
			DSN:             dsn,
			MaxOpenConns:    util.IntEnv(prefix+"MAX_OPEN_CONNS", 25),
			MaxIdleConns:    util.IntEnv(prefix+"MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: util.DurationEnv(prefix+"CONN_MAX_LIFETIME", 25),
			AutoMigrate:     util.BoolEnv(prefix+"AUTO_MIGRATE", false),
		})
	}

	return databases, nil
}

func defaultDBDriver(name string) string {
	if name == "primary" {
		return "postgres"
	}
	return ""
}
func defaultDBDSN(name string, driver string) string {
	if name == "primary" && driver == "postgres" {
		return "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable"
	}
	return ""
}
