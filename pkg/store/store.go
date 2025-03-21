package store

import (
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v70/github"
)

var (
	// Drivers defines the list of registered database drivers.
	Drivers = make(map[string]driver, 0)
)

type driver func(dsn string, logger *slog.Logger) (Store, error)

// Store provides the interface for the store implementations.
type Store interface {
	// WorkflowRunEvent
	StoreWorkflowRunEvent(*github.WorkflowRunEvent) error
	GetWorkflowRuns(time.Duration) ([]*WorkflowRun, error)
	PruneWorkflowRuns(time.Duration) error

	// WorkflowJobEvent
	StoreWorkflowJobEvent(*github.WorkflowJobEvent) error
	GetWorkflowJobs(time.Duration) ([]*WorkflowJob, error)
	PruneWorkflowJobs(time.Duration) error

	Open() (bool, error)
	Close() error
	Ping() (bool, error)
	Migrate() error
}

// New initializes a new database driver supported by current os.
func New(dsn string, logger *slog.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	if val, ok := Drivers[parsed.Scheme]; ok {
		return val(dsn, logger)
	}

	return nil, fmt.Errorf(
		"unknown database driver %s. available drivers are %v",
		parsed.Scheme,
		strings.Join(
			slices.Collect(
				maps.Keys(
					Drivers,
				),
			),
			", ",
		),
	)
}

func register(name string, f driver) {
	Drivers[name] = f
}
