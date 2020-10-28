package command

import (
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/promhippie/github_exporter/pkg/action"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/version"
	"github.com/urfave/cli/v2"
)

var (
	// ErrMissingGithubToken defines the error if github.token is empty.
	ErrMissingGithubToken = `Missing required github.token`
)

// Run parses the command line arguments and executes the program.
func Run() error {
	cfg := config.Load()

	app := &cli.App{
		Name:    "github_exporter",
		Version: version.String,
		Usage:   "GitHub Exporter",
		Authors: []*cli.Author{
			{
				Name:  "Thomas Boerger",
				Email: "thomas@webhippie.de",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log.level",
				Value:       "info",
				Usage:       "Only log messages with given severity",
				EnvVars:     []string{"GITHUB_EXPORTER_LOG_LEVEL"},
				Destination: &cfg.Logs.Level,
			},
			&cli.BoolFlag{
				Name:        "log.pretty",
				Value:       false,
				Usage:       "Enable pretty messages for logging",
				EnvVars:     []string{"GITHUB_EXPORTER_LOG_PRETTY"},
				Destination: &cfg.Logs.Pretty,
			},
			&cli.StringFlag{
				Name:        "web.address",
				Value:       "0.0.0.0:9504",
				Usage:       "Address to bind the metrics server",
				EnvVars:     []string{"GITHUB_EXPORTER_WEB_ADDRESS"},
				Destination: &cfg.Server.Addr,
			},
			&cli.StringFlag{
				Name:        "web.path",
				Value:       "/metrics",
				Usage:       "Path to bind the metrics server",
				EnvVars:     []string{"GITHUB_EXPORTER_WEB_PATH"},
				Destination: &cfg.Server.Path,
			},
			&cli.DurationFlag{
				Name:        "request.timeout",
				Value:       5 * time.Second,
				Usage:       "Request timeout as duration",
				EnvVars:     []string{"GITHUB_EXPORTER_REQUEST_TIMEOUT"},
				Destination: &cfg.Target.Timeout,
			},
			&cli.StringFlag{
				Name:        "github.token",
				Value:       "",
				Usage:       "Access token for the GitHub API",
				EnvVars:     []string{"GITHUB_EXPORTER_TOKEN"},
				Destination: &cfg.Target.Token,
			},
			&cli.StringFlag{
				Name:        "github.baseurl",
				Value:       "https://api.github.com/",
				Usage:       "URL to access the GitHub API",
				EnvVars:     []string{"GITHUB_EXPORTER_BASE_URL"},
				Destination: &cfg.Target.BaseURL,
			},
			&cli.StringSliceFlag{
				Name:        "github.enterprise",
				Value:       cli.NewStringSlice(),
				Usage:       "Enterprises to scrape metrics from",
				EnvVars:     []string{"GITHUB_EXPORTER_ENTERPRISE", "GITHUB_EXPORTER_ENTERPRISES"},
				Destination: &cfg.Target.Enterprises,
			},
			&cli.StringSliceFlag{
				Name:        "github.org",
				Value:       cli.NewStringSlice(),
				Usage:       "Organizations to scrape metrics from",
				EnvVars:     []string{"GITHUB_EXPORTER_ORG", "GITHUB_EXPORTER_ORGS"},
				Destination: &cfg.Target.Orgs,
			},
			&cli.StringSliceFlag{
				Name:        "github.repo",
				Value:       cli.NewStringSlice(),
				Usage:       "Repositories to scrape metrics from",
				EnvVars:     []string{"GITHUB_EXPORTER_REPO", "GITHUB_EXPORTER_REPOS"},
				Destination: &cfg.Target.Repos,
			},
			&cli.BoolFlag{
				Name:        "collector.orgs",
				Value:       true,
				Usage:       "Enable collector for orgs",
				EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_ORGS"},
				Destination: &cfg.Collector.Orgs,
			},
			&cli.BoolFlag{
				Name:        "collector.repos",
				Value:       true,
				Usage:       "Enable collector for repos",
				EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_REPOS"},
				Destination: &cfg.Collector.Repos,
			},
			&cli.BoolFlag{
				Name:        "collector.actions",
				Value:       false,
				Usage:       "Enable collector for actions",
				EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_ACTIONS"},
				Destination: &cfg.Collector.Actions,
			},
			&cli.BoolFlag{
				Name:        "collector.packages",
				Value:       false,
				Usage:       "Enable collector for packages",
				EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_PACKAGES"},
				Destination: &cfg.Collector.Packages,
			},
			&cli.BoolFlag{
				Name:        "collector.storage",
				Value:       false,
				Usage:       "Enable collector for storage",
				EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_STORAGE"},
				Destination: &cfg.Collector.Storage,
			},
		},
		Commands: []*cli.Command{
			Health(cfg),
		},
		Action: func(c *cli.Context) error {
			logger := setupLogger(cfg)

			if cfg.Target.Token == "" {
				level.Error(logger).Log(
					"msg", ErrMissingGithubToken,
				)

				return fmt.Errorf(ErrMissingGithubToken)
			}

			return action.Server(cfg, logger)
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Show the help, so what you see now",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Print the current version of that tool",
	}

	return app.Run(os.Args)
}
