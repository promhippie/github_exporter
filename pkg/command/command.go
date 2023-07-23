package command

import (
	"fmt"
	"os"
	"time"

	"github.com/promhippie/github_exporter/pkg/action"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/version"
	"github.com/urfave/cli/v2"
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
		Flags: RootFlags(cfg),
		Commands: []*cli.Command{
			Health(cfg),
		},
		Action: func(c *cli.Context) error {
			logger := setupLogger(cfg)

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

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return err
	}

	return nil
}

// RootFlags defines the available root flags.
func RootFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
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
		&cli.BoolFlag{
			Name:        "web.debug",
			Value:       false,
			Usage:       "Enable pprof debugging for server",
			EnvVars:     []string{"GITHUB_EXPORTER_WEB_PPROF"},
			Destination: &cfg.Server.Pprof,
		},
		&cli.DurationFlag{
			Name:        "web.timeout",
			Value:       10 * time.Second,
			Usage:       "Server metrics endpoint timeout",
			EnvVars:     []string{"GITHUB_EXPORTER_WEB_TIMEOUT"},
			Destination: &cfg.Server.Timeout,
		},
		&cli.StringFlag{
			Name:        "web.config",
			Value:       "",
			Usage:       "Path to web-config file",
			EnvVars:     []string{"GITHUB_EXPORTER_WEB_CONFIG"},
			Destination: &cfg.Server.Web,
		},
		&cli.DurationFlag{
			Name:        "request.timeout",
			Value:       5 * time.Second,
			Usage:       "Timeout requesting GitHub API",
			EnvVars:     []string{"GITHUB_EXPORTER_REQUEST_TIMEOUT"},
			Destination: &cfg.Target.Timeout,
		},
		&cli.StringFlag{
			Name:        "github.token",
			Value:       "",
			Usage:       "Access token for the GitHub API, string or path",
			EnvVars:     []string{"GITHUB_EXPORTER_TOKEN"},
			Destination: &cfg.Target.Token,
			Action: func(ctx *cli.Context, v string) error {
				if _, err := os.Stat(v); err == nil {
					content, err := os.ReadFile(v)
					if err != nil {
						return fmt.Errorf("Error reading token file:\n%w", err)
					}
					cfg.Target.Token = string(content)
				} else if os.IsNotExist(err) {
					cfg.Target.Token = v
				} else {
					return fmt.Errorf("Error reading token file:\n%w", err)
				}

				return nil
			},
		},
		&cli.Int64Flag{
			Name:        "github.app_id",
			Usage:       "App ID for the GitHub app",
			EnvVars:     []string{"GITHUB_EXPORTER_APP_ID"},
			Destination: &cfg.Target.AppID,
		},
		&cli.Int64Flag{
			Name:        "github.installation_id",
			Usage:       "Installation ID for the GitHub app",
			EnvVars:     []string{"GITHUB_EXPORTER_INSTALLATION_ID"},
			Destination: &cfg.Target.InstallID,
		},
		&cli.StringFlag{
			Name:        "github.private_key",
			Value:       "",
			Usage:       "Private key for the GitHub app, path or base64-encoded",
			EnvVars:     []string{"GITHUB_EXPORTER_PRIVATE_KEY"},
			Destination: &cfg.Target.PrivateKey,
		},
		&cli.StringFlag{
			Name:        "github.baseurl",
			Value:       "",
			Usage:       "URL to access the GitHub Enterprise API",
			EnvVars:     []string{"GITHUB_EXPORTER_BASE_URL"},
			Destination: &cfg.Target.BaseURL,
		},
		&cli.BoolFlag{
			Name:        "github.insecure",
			Value:       false,
			Usage:       "Skip TLS verification for GitHub Enterprise",
			EnvVars:     []string{"GITHUB_EXPORTER_INSECURE"},
			Destination: &cfg.Target.Insecure,
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
		&cli.IntFlag{
			Name:        "github.per-page",
			Value:       500,
			Usage:       "Number of records per page for API requests",
			EnvVars:     []string{"GITHUB_EXPORTER_PER_PAGE"},
			Destination: &cfg.Target.PerPage,
		},
		&cli.BoolFlag{
			Name:        "collector.admin",
			Value:       false,
			Usage:       "Enable collector for admin stats",
			EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_ADMIN"},
			Destination: &cfg.Collector.Admin,
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
			Name:        "collector.billing",
			Value:       false,
			Usage:       "Enable collector for billing",
			EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_BILLING"},
			Destination: &cfg.Collector.Billing,
		},
		&cli.BoolFlag{
			Name:        "collector.workflows",
			Value:       false,
			Usage:       "Enable collector for workflows",
			EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_WORKFLOWS"},
			Destination: &cfg.Collector.Workflows,
		},
		&cli.StringFlag{
			Name:        "collector.workflows.status",
			Value:       "",
			Usage:       "Query workflows with specific status",
			EnvVars:     []string{"GITHUB_EXPORTER_WORKFLOWS_STATUS"},
			Destination: &cfg.Target.Workflows.Status,
		},
		&cli.DurationFlag{
			Name:        "collector.workflows.window",
			Value:       12 * time.Hour,
			Usage:       "History window for querying workflows",
			EnvVars:     []string{"GITHUB_EXPORTER_WORKFLOWS_WINDOW"},
			Destination: &cfg.Target.Workflows.Window,
		},
		&cli.BoolFlag{
			Name:        "collector.runners",
			Value:       false,
			Usage:       "Enable collector for runners",
			EnvVars:     []string{"GITHUB_EXPORTER_COLLECTOR_RUNNERS"},
			Destination: &cfg.Collector.Runners,
		},
	}
}
