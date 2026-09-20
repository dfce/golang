package logging

import (
	"fmt"
	"go-module/internal/config"
	"go-module/pkg/constant"
	"go-module/pkg/util"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// KV2Fields 接收交替的 Key-Value 健值对，转换为 []zap.Field
// 例： KV2Fields("ENV", "dev", "app", "my-app")
func KV2Fields(kvs ...any) []zap.Field {
	num := len(kvs)
	if num == 0 {
		return nil
	}
	// 预分配内存，容量为 1/2 kvs
	fields := make([]zap.Field, 0, num/2)

	for i := 0; i < num; i += 2 {
		// 确保不被越界 (一定是偶数个数参数)
		if i+1 >= num {
			// fields = append(fields, zap.Any(fmt.Sprintf("KEY_WITHOUT_VALUE_%d", i), kvs[i]))
			// 直接结束， 不给赋值
			break
		}
		// key 必须是string
		keyStr, ok := kvs[i].(string)
		if !ok {
			keyStr = fmt.Sprintf("INVALID_KEY_TYPE_%d", i)
		}
		fields = append(fields, zap.Any(keyStr, kvs[i+1]))
	}
	return fields
}

func New(cfg *config.Config) *zap.Logger {
	isDev := util.IsDev(cfg.App.ENV)

	// 日志输出格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(constant.TimeFormat))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 🚀 关键：让 INFO/WARN/ERROR 变色

	var encoder zapcore.Encoder = zapcore.NewJSONEncoder(encoderConfig)
	if isDev {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	level := parseLevel(cfg.Log.Level, isDev)
	cores := getCores(encoder, level, cfg.Log.Mode, cfg.Log.Dir, cfg.App.Name)
	// Logger
	// zap.AddCaller() 可以让日志带上代码行号
	// logger := zap.New(zapcore.NewTee(core...), zap.AddCaller()
	logger := zap.New(zapcore.NewTee(cores...))
	zap.RedirectStdLog(logger) // 接管 Go 原生的 log.Printf

	return logger
}

func parseLevel(value string, isDev bool) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	case "":
		if isDev {
			return zap.DebugLevel
		}
		return zap.InfoLevel
	default:
		return zap.InfoLevel
	}
}

func getCores(encoder zapcore.Encoder, level zapcore.Level, mode, dir, appName string) []zapcore.Core {
	stdout := zapcore.NewCore(encoder.Clone(), zapcore.AddSync(os.Stdout), level)
	file := zapcore.NewCore(encoder.Clone(), zapcore.AddSync(getLumber(dir, appName)), level)
	switch mode {
	case "0":
		return []zapcore.Core{stdout, file}
	case "1":
		return []zapcore.Core{file}
	case "2":
		return []zapcore.Core{stdout}
	default:
		return []zapcore.Core{stdout}
	}
}

func getLumber(dir, appName string) *lumberjack.Logger {
	_ = os.MkdirAll(dir, os.ModePerm)
	// 日志轮转切割
	return &lumberjack.Logger{
		Filename:   filepath.Join(dir, appName+".log"), // 日志文件路径
		MaxSize:    100,                                // 单个日志文件最大大小（单位：MB）。超过切割
		MaxBackups: 30,                                 // 保留旧日志文件个数
		MaxAge:     7,                                  // 保留旧日志文件最大天数（按天切割的核心保障）
		Compress:   true,                               // 是否压缩/gzip 旧日志文件
		LocalTime:  true,                               // 使用本地时间命名备份文件
	}
}
