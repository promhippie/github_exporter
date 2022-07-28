package command

import (
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/promhippie/github_exporter/pkg/config"
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
