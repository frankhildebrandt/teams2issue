package observability

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/frankhildebrandt/teams2issue/internal/config"
)

func NewLogger(cfg config.Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Logging.Level)
	if err != nil {
		return nil, err
	}

	options := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.Logging.AddSource,
	}

	var handler slog.Handler
	switch cfg.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, options)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, options)
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Logging.Format)
	}

	return slog.New(handler), nil
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", value)
	}
}
