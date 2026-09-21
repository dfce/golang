package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"go-module/internal/config"
	"go-module/internal/modules"
	"go-module/internal/platform/datastore"
	"go-module/pkg/jwt"
	"go-module/pkg/util"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
	dbs    datastore.Databases
	redis  *datastore.RedisSvc
}

func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	var tokenService *jwt.Service
	var err error
	tokenService, err = jwt.NewService(jwt.Config{
		SecretKey: cfg.Auth.SecretKey,
		Expiry:    cfg.Auth.TokenExpiry,
		Issuer:    cfg.Auth.Issuer,
	})
	if err != nil {
		return nil, fmt.Errorf("configure JWT service: %w", err)
	}

	dbs, err := datastore.OpenDatabases(cfg.Databases, logger, util.IsDev(cfg.App.ENV))
	if err != nil {
		return nil, err
	}

	redisClient, err := datastore.OpenRedis(ctx, cfg.Redis, logger)
	if err != nil {
		datastore.CloseDatabases(dbs, logger)
		return nil, fmt.Errorf("connect Redis: %w", err)
	}

	router := modules.BuildAppEngine(cfg, dbs, redisClient, logger, tokenService)
	return &App{
		cfg:    cfg,
		logger: logger,
		server: &http.Server{
			Addr:              cfg.HTTP.Addr(),
			Handler:           router,
			ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
			WriteTimeout:      cfg.HTTP.WriteTimeout,
			IdleTimeout:       cfg.HTTP.IdleTimeout,
		},
		dbs:   dbs,
		redis: redisClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		runmsg := fmt.Sprintf("http server listening addr: %s, env: %s", a.server.Addr, a.cfg.App.ENV)
		a.logger.Info(runmsg)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 先关闭 HTTP 服务（拒绝新请求，等待存量请求处理结束）
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("HTTP 服务优雅关闭失败(可能超时强制退出)", zap.Any("shutdownError", err))
			a.close()
			return err
		}

		a.logger.Debug("HTTP 服务已安全停止，当前无进行中的请求")
		a.close()
		return nil

	case err := <-errCh:
		a.close()
		return err
	}
}

func (a *App) close() {
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.logger.Warn("close redis failed", zap.Any("error", err))
		}
	}
	datastore.CloseDatabases(a.dbs, a.logger)
}
