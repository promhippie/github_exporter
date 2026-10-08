package command

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
	"go.uber.org/automaxprocs/maxprocs"
)

func setupLogger(cfg *config.Config) *slog.Logger {
	if cfg.Logs.Pretty {
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: loggerLevel(cfg),
			}),
		)
	}

	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: loggerLevel(cfg),
		}),
	)
}

func loggerLevel(cfg *config.Config) slog.Leveler {
	switch strings.ToLower(cfg.Logs.Level) {
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func setupMaxProcs(logger *slog.Logger) {
	if _, err := maxprocs.Set(maxprocs.Logger(func(format string, args ...any) {
		logger.Info(fmt.Sprintf(format, args...))
	})); err != nil {
		logger.Warn("Failed to set GOMAXPROCS",
			"error", err,
		)
	}
}

func setupStorage(cfg *config.Config, logger *slog.Logger) (store.Store, error) {
	dsn, err := config.Value(cfg.Database.DSN)

	if err != nil {
		return nil, fmt.Errorf("failed to read dsn: %w", err)
	}

	return store.New(dsn, logger)
}
