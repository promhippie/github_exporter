package command

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

func setupLogger(cfg *config.Config) log.Logger {
	var logger log.Logger

	if cfg.Logs.Pretty {
		logger = log.NewSyncLogger(
			log.NewLogfmtLogger(os.Stdout),
		)
	} else {
		logger = log.NewSyncLogger(
			log.NewJSONLogger(os.Stdout),
		)
	}

	switch strings.ToLower(cfg.Logs.Level) {
	case "error":
		logger = level.NewFilter(logger, level.AllowError())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	default:
		logger = level.NewFilter(logger, level.AllowInfo())
	}

	return log.With(
		logger,
		"ts", log.DefaultTimestampUTC,
	)
}

func setupStorage(cfg *config.Config, logger log.Logger) (store.Store, error) {
	dsn, err := config.Value(cfg.Database.DSN)

	if err != nil {
		return nil, fmt.Errorf("failed to read dsn: %w", err)
	}

	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	switch parsed.Scheme {
	case "sqlite", "sqlite3":
		return store.NewGenericStore(dsn, logger)
	case "mysql", "mariadb":
		return store.NewGenericStore(dsn, logger)
	case "postgres", "postgresql":
		return store.NewGenericStore(dsn, logger)
	}

	return nil, store.ErrUnknownDriver
}
