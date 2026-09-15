package logging

import (
	"generatego/pkg/constant"
	"generatego/pkg/util"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 配置默认参数
var (
	logDir  = util.GetEnv("LOG_DIR", "logs")
	logFile = util.GetEnv("APP_NAME", "gen-app") + ".log"
	logMod  = util.GetEnv("LOG_MOD", "0") // 日志输出 0-两者都 1-文件 2-控制台
	isDev   = util.IsDev(util.GetEnv("ENV", "dev"))
)

func New() *zap.Logger {

	// 2. 定制日志输出格式(Encoder)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(constant.TimeFormat))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 大写日志级别（INFO, ERROR）

	// 创建 JSON 格式（生产环境）或 Console 格式
	var encoder zapcore.Encoder = zapcore.NewJSONEncoder(encoderConfig)
	if isDev {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 3. 配置多路输出（Core）
	var zapLogLev zapcore.Level = zap.DebugLevel
	if !isDev {
		zapLogLev = zap.InfoLevel
	}
	// 开发环境：同时输出到控制台、文件
	// 生产环境：只写入文件
	// 若有日志采集环境（k8s docker ）只输出控制台即可；否则输出到指定文件
	core := getCores(encoder, zapLogLev)

	// 4. Logger
	// zap.AddCaller() 可以让日志带上代码行号
	// logger := zap.New(zapcore.NewTee(core...), zap.AddCaller())
	logger := zap.New(zapcore.NewTee(core...))
	zap.RedirectStdLog(logger) // 接管 Go 原生的 log.Printf

	return logger
}

func getCores(encoder zapcore.Encoder, level zapcore.Level) []zapcore.Core {
	stdout := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	file := zapcore.NewCore(encoder, zapcore.AddSync(getLumber()), level)
	switch logMod {
	case "0":
		return []zapcore.Core{stdout, file}
	case "1":
		return []zapcore.Core{file}
	case "2":
		return []zapcore.Core{stdout}
	default:
		return nil
	}
}

func getLumber() *lumberjack.Logger {
	if logMod > "1" {
		return &lumberjack.Logger{}
	}
	// 创建日志存放目录
	_ = os.MkdirAll(logDir, os.ModePerm)
	// 1. 配置 Lumberjack 实现日志轮转切割
	return &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFile), // 日志文件路径
		MaxSize:    100,                            // 单个日志文件最大大小（单位：MB）。超过切割
		MaxBackups: 30,                             // 保留旧日志文件个数
		MaxAge:     7,                              // 保留旧日志文件最大天数（按天切割的核心保障）
		Compress:   true,                           // 是否压缩/gzip 旧日志文件
		LocalTime:  true,                           // 使用本地时间命名备份文件
	}
}
