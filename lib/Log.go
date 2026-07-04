package lib

import (
	"config"
	"log/slog"
	"os"
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
