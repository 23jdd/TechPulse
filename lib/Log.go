package lib

import (
	"config"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *slog.Logger

func InitLogger(config config.Config) {
	if config.Mode == "dev" {
		Log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		).With(
			"app", "TechPulse",
			"env", "dev",
		)
	} else {
		logFile := &lumberjack.Logger{
			Filename:   config.Path, // 当前日志文件
			MaxSize:    10,          // 单文件最大 10MB
			MaxBackups: 5,           // 保留 5 个旧文件
			MaxAge:     7,           // 保留 7 天
			Compress:   true,        // 是否压缩 gzip
		}

		logger := slog.New(
			slog.NewJSONHandler(logFile, nil),
		)

		slog.SetDefault(logger)

		slog.Info("server start")
		slog.Info("user login", "user_id", 123)

		Log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		).With(
			"app", "TechPulse",
			"env", "release",
		)
	}
}
