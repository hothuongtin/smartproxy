package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/MatusOllah/slogcolor"
)

// Config represents logging configuration
type Config struct {
	Level string `yaml:"level"`
}

// SetupLogger configures slog with colored output using slogcolor
func SetupLogger(config *Config) *slog.Logger {
	// Determine log level from config
	level := slog.LevelInfo
	if config != nil && config.Level != "" {
		switch strings.ToLower(config.Level) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn", "warning":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}

	opts := &slogcolor.Options{
		Level:         level,
		TimeFormat:    "15:04:05.000",
		SrcFileMode:   slogcolor.ShortFile,
		SrcFileLength: 0,
		MsgPrefix:     "",
	}

	// Show more detailed source info in debug mode
	if level == slog.LevelDebug {
		opts.SrcFileMode = slogcolor.LongFile
	}

	handler := slogcolor.NewHandler(os.Stdout, opts)
	return slog.New(handler)
}
