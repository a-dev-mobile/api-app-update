package logging

import (
	"github.com/a-dev-mobile/api-app-update/internal/config"
	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"

	"os"
)

func SetupLogger(cfg *config.Config) *slog.Logger {
	level := parseLogLevel(cfg.Logging.Level)
	var logger *slog.Logger

	// Setup lumberjack for log rotation
	logWriter := &lumberjack.Logger{
		Filename:   cfg.Logging.FileOutput.FilePath,
		MaxSize:    cfg.Logging.FileOutput.MaxSizeMB,
		MaxBackups: cfg.Logging.FileOutput.MaxBackups,
		MaxAge:     parseMaxAge(cfg.Logging.FileOutput.RotationPolicy),
		Compress:   true,
	}

	// Use lumberjack for file logging if file path is specified
	if cfg.Logging.FileOutput.FilePath != "" {
		logger = slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: level}))
	} else {
		// Use standard output if file path is not specified
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	return logger
}

func parseLogLevel(level config.LogLevel) slog.Level {

	switch level {
	case config.LogLevelInfo:
		return slog.LevelInfo
	case config.LogLevelWarning:
		return slog.LevelWarn
	case config.LogLevelError:
		return slog.LevelError
	case config.LogLevelDebug:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// parseMaxAge converts rotation policy into max age in days.
func parseMaxAge(rotationPolicy config.RotationPolicy) int {
	switch rotationPolicy {
	case config.Daily:
		return 1 // 1 day for daily rotation
	case config.Weekly:
		return 7 // 7 days for weekly rotation
	case config.Monthly:
		return 30 // 30 days for monthly rotation
	default:
		return 0 // No log rotation
	}
}
