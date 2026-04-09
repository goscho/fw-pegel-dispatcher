package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// New returns a slog.Logger writing to stdout and a rolling file under logDir/logs/app.log.
func New(logDir string) (*slog.Logger, error) {
	logsDir := filepath.Join(logDir, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, err
	}
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logsDir, "app.log"),
		MaxSize:    50,
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	}
	mw := io.MultiWriter(os.Stdout, fileWriter)
	opts := &slog.HandlerOptions{
		Level: parseLevel(os.Getenv("LOG_LEVEL")),
	}
	h := slog.NewTextHandler(mw, opts)
	return slog.New(h), nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
