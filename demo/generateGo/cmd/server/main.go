package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"generatego/internal/bootstrap"
	"generatego/internal/config"
	"generatego/internal/platform/logging"

	"go.uber.org/zap"
)

/*
httpServer 入口主函数
*/
// @title 为服务 API 接口文档
// @version 1.0
// @description 基于 Cin + GORM + Redis 封装的高性能后端脚手架
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger := logging.New()
	// 【核心：shutdown收尾】在main函数彻底退出时， 强行将内存中未落盘的日志刷写到磁盘上
	defer func() {
		// 忽略处理 logger.Sync .Err
		_ = logger.Sync()
	}()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config error", zap.Error(err))
		os.Exit(1)
	}
	logger.Debug("load config: ", zap.Any("config", cfg))

	// appRun
	app, err := bootstrap.New(ctx, cfg, logger)
	if err := app.Run(ctx); err != nil {
		logger.Error("server stopped with error", zap.Any("error", err))
		os.Exit(1)
	}

}
