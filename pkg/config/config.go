package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// Server defines the general server configuration.
type Server struct {
	Addr    string
	Path    string
	Timeout time.Duration
	Web     string
	Pprof   bool
}

// Webhook defines the webhook specific configuration.
type Webhook struct {
	Path   string
	Secret string
}

// Logs defines the level and color for log configuration.
type Logs struct {
	Level  string
	Pretty bool
}

// WorkflowRuns defines the workflow run specific configuration.
type WorkflowRuns struct {
	Window      time.Duration
	PurgeWindow time.Duration
	Labels      cli.StringSlice
}

// WorkflowJobs defines the workflow job specific configuration.
type WorkflowJobs struct {
	Window      time.Duration
	PurgeWindow time.Duration
	Labels      cli.StringSlice
}

// Runners defines the runner specific configuration.
type Runners struct {
	Labels cli.StringSlice
}

// Target defines the target specific configuration.
type Target struct {
	Token        string
	PrivateKey   string
	AppID        int64
	InstallID    int64
	BaseURL      string
	Insecure     bool
	Enterprises  cli.StringSlice
	Orgs         cli.StringSlice
	Repos        cli.StringSlice
	Timeout      time.Duration
	PerPage      int
	WorkflowRuns WorkflowRuns
	WorkflowJobs WorkflowJobs
	Runners      Runners
}

// Collector defines the collector specific configuration.
type Collector struct {
	Admin        bool
	Orgs         bool
	Repos        bool
	Billing      bool
	WorkflowRuns bool
	WorkflowJobs bool
	Runners      bool
}

// Database defines the database specific configuration.
type Database struct {
	DSN string
}

// Config is a combination of all available configurations.
type Config struct {
	Server    Server
	Webhook   Webhook
	Logs      Logs
	Target    Target
	Collector Collector
	Database  Database
}

// Load initializes a default configuration struct.
func Load() *Config {
	return &Config{}
}

// RunLabels defines the default labels used by workflow run collector.
func RunLabels() *cli.StringSlice {
	return cli.NewStringSlice(
		"owner",
		"repo",
		"workflow",
		"event",
		"name",
		"status",
		"branch",
		"number",
		"run",
	)
}

// JobLabels defines the default labels used by workflow job collector.
func JobLabels() *cli.StringSlice {
	return cli.NewStringSlice(
		"owner",
		"repo",
		"name",
		"title",
		"branch",
		"sha",
		"identifier",
		"run_id",
		"run_attempt",
		"labels",
		"runner_id",
		"runner_name",
		"runner_group_id",
		"runner_group_name",
		"workflow_name",
	)
}

// RunnerLabels defines the default labels used by runner collector.
func RunnerLabels() *cli.StringSlice {
	return cli.NewStringSlice(
		"owner",
		"id",
		"name",
		"os",
		"status",
	)
}

// Value returns the config value based on a DSN.
func Value(val string) (string, error) {
	if strings.HasPrefix(val, "file://") {
		content, err := os.ReadFile(
			strings.TrimPrefix(val, "file://"),
		)

		if err != nil {
			return "", fmt.Errorf("failed to parse secret file: %w", err)
		}

		return string(content), nil
	}

	if strings.HasPrefix(val, "base64://") {
		content, err := base64.StdEncoding.DecodeString(
			strings.TrimPrefix(val, "base64://"),
		)

		if err != nil {
			return "", fmt.Errorf("failed to parse base64 value: %w", err)
		}

		return string(content), nil
	}

	return val, nil
}
