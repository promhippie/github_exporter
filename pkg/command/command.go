package command

import (
	"context"
	"os"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/promhippie/github_exporter/pkg/action"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
	"github.com/promhippie/github_exporter/pkg/version"
	"github.com/urfave/cli/v3"
)

var (
	defaultDatabaseDSN = ""
)

func init() {
	if _, ok := store.Drivers["chai"]; ok {
		defaultDatabaseDSN = "chai://storage/exporter"
	} else if _, ok := store.Drivers["sqlite"]; ok {
		defaultDatabaseDSN = "sqlite://storage/exporter.sqlite3"
	}
}

// Run parses the command line arguments and executes the program.
func Run() error {
	cfg := config.Load()

	app := &cli.Command{
		Name:    "github_exporter",
		Version: version.String,
		Usage:   "GitHub Exporter",
		Authors: []any{
			"Thomas Boerger <thomas@webhippie.de>",
		},
		Flags: RootFlags(cfg),
		Commands: []*cli.Command{
			Health(cfg),
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			logger := setupLogger(cfg)
			db, err := setupStorage(cfg, logger)

			if err != nil {
				logger.Error("Failed to setup database",
					"error", err,
				)

				return err
			}

			if db != nil {
				defer db.Close()
			}

			if _, err := backoff.Retry(
				ctx,
				db.Open,
				backoff.WithBackOff(backoff.NewExponentialBackOff()),
				backoff.WithNotify(func(err error, dur time.Duration) {
					logger.Warn("Database open failed",
						"retry", dur,
						"error", err,
					)
				}),
			); err != nil {
				logger.Error("Giving up to connect to database",
					"error", err,
				)

				return err
			}

			if _, err := backoff.Retry(
				ctx,
				db.Ping,
				backoff.WithBackOff(backoff.NewExponentialBackOff()),
				backoff.WithNotify(func(err error, dur time.Duration) {
					logger.Warn("Database ping failed",
						"retry", dur,
						"err", err,
					)
				}),
			); err != nil {
				logger.Error("Giving up to ping the database",
					"error", err,
				)

				return err
			}

			if err := db.Migrate(); err != nil {
				logger.Error("Failed to migrate database",
					"error", err,
				)
			}

			if cfg.Target.WorkflowRuns.PurgeWindow < cfg.Target.WorkflowRuns.Window {
				logger.Warn("Workflow Run purge window cannot be smaller than query window or data loss will occur", "config", cfg.Target.WorkflowRuns)
			}
			if cfg.Target.WorkflowJobs.PurgeWindow < cfg.Target.WorkflowJobs.Window {
				logger.Warn("Workflow Run purge window cannot be smaller than query window or data loss will occur", "config", cfg.Target.WorkflowJobs)
			}

			return action.Server(cfg, db, logger)
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

	return app.Run(context.Background(), os.Args)
}

// RootFlags defines the available root flags.
func RootFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log.level",
			Value:       "info",
			Usage:       "Only log messages with given severity",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_LOG_LEVEL"),
			Destination: &cfg.Logs.Level,
		},
		&cli.BoolFlag{
			Name:        "log.pretty",
			Value:       false,
			Usage:       "Enable pretty messages for logging",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_LOG_PRETTY"),
			Destination: &cfg.Logs.Pretty,
		},
		&cli.StringFlag{
			Name:        "web.address",
			Value:       "0.0.0.0:9504",
			Usage:       "Address to bind the metrics server",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEB_ADDRESS"),
			Destination: &cfg.Server.Addr,
		},
		&cli.StringFlag{
			Name:        "web.path",
			Value:       "/metrics",
			Usage:       "Path to bind the metrics server",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEB_PATH"),
			Destination: &cfg.Server.Path,
		},
		&cli.BoolFlag{
			Name:        "web.debug",
			Value:       false,
			Usage:       "Enable pprof debugging for server",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEB_PPROF"),
			Destination: &cfg.Server.Pprof,
		},
		&cli.DurationFlag{
			Name:        "web.timeout",
			Value:       10 * time.Second,
			Usage:       "Server metrics endpoint timeout",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEB_TIMEOUT"),
			Destination: &cfg.Server.Timeout,
		},
		&cli.StringFlag{
			Name:        "web.config",
			Value:       "",
			Usage:       "Path to web-config file",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEB_CONFIG"),
			Destination: &cfg.Server.Web,
		},
		&cli.StringFlag{
			Name:        "webhook.path",
			Value:       "/github",
			Usage:       "Path to webhook target for GitHub",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEBHOOK_PATH"),
			Destination: &cfg.Webhook.Path,
		},
		&cli.StringFlag{
			Name:        "webhook.secret",
			Value:       "",
			Usage:       "Secret used by GitHub to access webhook",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WEBHOOK_SECRET"),
			Destination: &cfg.Webhook.Secret,
		},
		&cli.StringFlag{
			Name:        "database.dsn",
			Value:       defaultDatabaseDSN,
			Usage:       "DSN for the database connection",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_DATABASE_DSN"),
			Destination: &cfg.Database.DSN,
		},
		&cli.DurationFlag{
			Name:        "request.timeout",
			Value:       5 * time.Second,
			Usage:       "Timeout requesting GitHub API",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_REQUEST_TIMEOUT"),
			Destination: &cfg.Target.Timeout,
		},
		&cli.StringFlag{
			Name:        "github.token",
			Value:       "",
			Usage:       "Access token for the GitHub API, also supports file:// and base64://",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_TOKEN"),
			Destination: &cfg.Target.Token,
		},
		&cli.Int64Flag{
			Name:        "github.app_id",
			Usage:       "App ID for the GitHub app",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_APP_ID"),
			Destination: &cfg.Target.AppID,
		},
		&cli.Int64Flag{
			Name:        "github.installation_id",
			Usage:       "Installation ID for the GitHub app",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_INSTALLATION_ID"),
			Destination: &cfg.Target.InstallID,
		},
		&cli.StringFlag{
			Name:        "github.private_key",
			Value:       "",
			Usage:       "Private key for the GitHub app, also supports file:// and base64://",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_PRIVATE_KEY"),
			Destination: &cfg.Target.PrivateKey,
		},
		&cli.StringFlag{
			Name:        "github.baseurl",
			Value:       "",
			Usage:       "URL to access the GitHub Enterprise API",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_BASE_URL"),
			Destination: &cfg.Target.BaseURL,
		},
		&cli.BoolFlag{
			Name:        "github.insecure",
			Value:       false,
			Usage:       "Skip TLS verification for GitHub Enterprise",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_INSECURE"),
			Destination: &cfg.Target.Insecure,
		},
		&cli.StringSliceFlag{
			Name:        "github.enterprise",
			Value:       []string{},
			Usage:       "Enterprises to scrape metrics from",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_ENTERPRISE", "GITHUB_EXPORTER_ENTERPRISES"),
			Destination: &cfg.Target.Enterprises,
		},
		&cli.StringSliceFlag{
			Name:        "github.org",
			Value:       []string{},
			Usage:       "Organizations to scrape metrics from",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_ORG", "GITHUB_EXPORTER_ORGS"),
			Destination: &cfg.Target.Orgs,
		},
		&cli.StringSliceFlag{
			Name:        "github.repo",
			Value:       []string{},
			Usage:       "Repositories to scrape metrics from",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_REPO", "GITHUB_EXPORTER_REPOS"),
			Destination: &cfg.Target.Repos,
		},
		&cli.IntFlag{
			Name:        "github.per-page",
			Value:       500,
			Usage:       "Number of records per page for API requests",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_PER_PAGE"),
			Destination: &cfg.Target.PerPage,
		},
		&cli.BoolFlag{
			Name:        "collector.admin",
			Value:       false,
			Usage:       "Enable collector for admin stats",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_ADMIN"),
			Destination: &cfg.Collector.Admin,
		},
		&cli.BoolFlag{
			Name:        "collector.orgs",
			Value:       true,
			Usage:       "Enable collector for orgs",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_ORGS"),
			Destination: &cfg.Collector.Orgs,
		},
		&cli.BoolFlag{
			Name:        "collector.repos",
			Value:       true,
			Usage:       "Enable collector for repos",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_REPOS"),
			Destination: &cfg.Collector.Repos,
		},
		&cli.BoolFlag{
			Name:        "collector.billing",
			Value:       false,
			Usage:       "Enable collector for billing",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_BILLING"),
			Destination: &cfg.Collector.Billing,
		},
		&cli.BoolFlag{
			Name:        "collector.workflow_runs",
			Value:       false,
			Usage:       "Enable collector for workflows",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_WORKFLOW_RUNS"),
			Destination: &cfg.Collector.WorkflowRuns,
		},
		&cli.DurationFlag{
			Name:        "collector.workflow_runs.window",
			Value:       24 * time.Hour,
			Usage:       "History window for querying workflows",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_RUNS_WINDOW"),
			Destination: &cfg.Target.WorkflowRuns.Window,
		},
		&cli.DurationFlag{
			Name:        "collector.workflow_runs.purge_window",
			Value:       24 * time.Hour,
			Usage:       "History window for keeping data in database. Defaults to the query window",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_RUNS_PURGE_WINDOW"),
			Destination: &cfg.Target.WorkflowRuns.PurgeWindow,
		},
		&cli.StringSliceFlag{
			Name:        "collector.workflow_runs.labels",
			Value:       config.RunLabels(),
			Usage:       "List of labels used for workflows",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_RUNS_LABELS"),
			Destination: &cfg.Target.WorkflowRuns.Labels,
		},
		&cli.BoolFlag{
			Name:        "collector.workflow_jobs",
			Value:       false,
			Usage:       "Enable collector for workflow jobs",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_WORKFLOW_JOBS"),
			Destination: &cfg.Collector.WorkflowJobs,
		},
		&cli.DurationFlag{
			Name:        "collector.workflow_jobs.window",
			Value:       24 * time.Hour,
			Usage:       "History window for querying workflow jobs",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_JOBS_WINDOW"),
			Destination: &cfg.Target.WorkflowJobs.Window,
		},
		&cli.DurationFlag{
			Name:        "collector.workflow_jobs.purge_window",
			Value:       24 * time.Hour,
			Usage:       "History window for keeping data in database. Defaults to the query window",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_JOBS_PURGE_WINDOW"),
			Destination: &cfg.Target.WorkflowJobs.PurgeWindow,
		},
		&cli.StringSliceFlag{
			Name:        "collector.workflow_jobs.labels",
			Value:       config.JobLabels(),
			Usage:       "List of labels used for workflow jobs",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_WORKFLOW_JOBS_LABELS"),
			Destination: &cfg.Target.WorkflowJobs.Labels,
		},
		&cli.BoolFlag{
			Name:        "collector.runners",
			Value:       false,
			Usage:       "Enable collector for runners",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_COLLECTOR_RUNNERS"),
			Destination: &cfg.Collector.Runners,
		},
		&cli.StringSliceFlag{
			Name:        "collector.runners.labels",
			Value:       config.RunnerLabels(),
			Usage:       "List of labels used for runners",
			Sources:     cli.EnvVars("GITHUB_EXPORTER_RUNNERS_LABELS"),
			Destination: &cfg.Target.Runners.Labels,
		},
	}
}
