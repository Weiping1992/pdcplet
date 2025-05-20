package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Level   string   `mapstructure:"level"`
	Format  string   `mapstructure:"format"`
	Outputs []string `mapstructure:"outputs"`
	File    struct {
		Path       string `mapstructure:"path"`
		MaxSize    int    `mapstructure:"maxSize"`
		Compress   bool   `mapstructure:"compress"`
		MaxBackups int    `mapstructure:"maxBackups"`
	} `mapstructure:"file"`
}

func InitLogger(cfg LogConfig, appName string, version string) {
	handler := createSlogHandler(cfg)
	logger := slog.New(handler).With(
		slog.String("app", appName),
		slog.String("version", version),
	)
	slog.SetDefault(logger)
}

func createSlogHandler(cfg LogConfig) slog.Handler {
	var outputs []io.Writer

	// 根据配置添加输出目标
	for _, out := range cfg.Outputs {
		switch out {
		case "stdout":
			outputs = append(outputs, os.Stdout)
		case "file":
			fileWriter := &lumberjack.Logger{
				Filename:   cfg.File.Path,
				MaxSize:    cfg.File.MaxSize,
				Compress:   cfg.File.Compress,
				MaxBackups: cfg.File.MaxBackups,
			}
			outputs = append(outputs, fileWriter)
		}
	}

	// 创建多目标写入器
	multiWriter := io.MultiWriter(outputs...)

	// 动态选择日志格式
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level:     parseLogLevel(cfg.Level),
			AddSource: true,
		})
	default: // text 格式
		handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: parseLogLevel(cfg.Level),
		})
	}

	return handler
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
