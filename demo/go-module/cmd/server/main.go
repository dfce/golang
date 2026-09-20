package main

import (
	"context"
	"fmt"
	"go-module/internal/bootstrap"
	"go-module/internal/config"
	"go-module/internal/platform/logging"
	_ "go-module/swdocs"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

/*
httpServer 入口主函数
*/
// @title 服务API接口文档
// @version 1.0
// @description 基于 Gin + GORM + Redis 搭建的模块化服务脚手架
// @BasePath /api/v1
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// 加载配置 ENV, 失败退出
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config error: %v\n", err)
		os.Exit(1)
	}
	// Logger init
	logger := logging.New(cfg)
	// 【核心：shutdown收尾】在main函数彻底退出时， 强行将内存中未落盘的日志刷写到磁盘上
	defer func() {
		// 忽略处理 logger.Sync .Err
		_ = logger.Sync()
	}()

	// logger.Debug("load config", zap.String("ENV", cfg.App.ENV), zap.String("app", cfg.App.Name))
	logger.Debug("load config", logging.KV2Fields("env", cfg.App.ENV, "app", cfg.App.Name)...)

	// appRun
	app, err := bootstrap.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("Build app error", zap.Error(err))
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		logger.Error("Server stoped with error", zap.Any("error", err))
		os.Exit(1)
	}
}
