package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/sanketpadhyal/towie/internal/config"
)

func New(cfg config.Logging) *slog.Logger {
	level := parseLevel(cfg.Level)
	out := parseOutput(cfg.Output)

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func parseOutput(output string) io.Writer {
	switch output {
	case "stderr":
		return os.Stderr
	case "stdout", "":
		return os.Stdout
	default:
		f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return os.Stdout
		}
		return f
	}
}
