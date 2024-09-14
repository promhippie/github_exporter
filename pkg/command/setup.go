package command

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
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

func setupStorage(cfg *config.Config, logger *slog.Logger) (store.Store, error) {
	dsn, err := config.Value(cfg.Database.DSN)

	if err != nil {
		return nil, fmt.Errorf("failed to read dsn: %w", err)
	}

	return store.New(dsn, logger)
}
