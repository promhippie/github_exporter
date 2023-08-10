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

// Workflows defines the workflow specific configuration.
type Workflows struct {
	Window time.Duration
	Labels cli.StringSlice
}

// Target defines the target specific configuration.
type Target struct {
	Token       string
	PrivateKey  string
	AppID       int64
	InstallID   int64
	BaseURL     string
	Insecure    bool
	Enterprises cli.StringSlice
	Orgs        cli.StringSlice
	Repos       cli.StringSlice
	Timeout     time.Duration
	PerPage     int
	Workflows   Workflows
}

// Collector defines the collector specific configuration.
type Collector struct {
	Admin     bool
	Orgs      bool
	Repos     bool
	Billing   bool
	Workflows bool
	Runners   bool
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

// Labels defines the default labels used by workflow collector.
func Labels() *cli.StringSlice {
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
