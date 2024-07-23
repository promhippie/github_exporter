package store

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/google/go-github/v63/github"
)

var (
	// ErrUnknownDriver defines a named error for unknown driver.
	ErrUnknownDriver = errors.New("unknown database driver")

	// Drivers defines the list of registered database drivers.
	Drivers = make(map[string]driver, 0)
)

type driver func(dsn string, logger log.Logger) (Store, error)

// Store provides the interface for the store implementations.
type Store interface {
	StoreWorkflowRunEvent(*github.WorkflowRunEvent) error
	GetWorkflowRuns() ([]*WorkflowRun, error)
	PruneWorkflowRuns(time.Duration) error

	Open() error
	Close() error
	Ping() error
	Migrate() error
}

// New initializes a new database driver supported by current os.
func New(dsn string, logger log.Logger) (Store, error) {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	if val, ok := Drivers[parsed.Scheme]; ok {
		return val(dsn, logger)
	}

	return nil, ErrUnknownDriver
}

func register(name string, f driver) {
	Drivers[name] = f
}
