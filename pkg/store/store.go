package store

import (
	"errors"
	"time"

	"github.com/google/go-github/v56/github"
)

var (
	// ErrUnknownDriver defines a named error for unknown store drivers.
	ErrUnknownDriver = errors.New("unknown database driver")
)

// Store provides the interface for the store implementations.
type Store interface {
	StoreWorkflowRunEvent(*github.WorkflowRunEvent) error
	CreateOrUpdateWorkflowRun(*WorkflowRun) error
	GetWorkflowRuns() ([]*WorkflowRun, error)
	PruneWorkflowRuns(time.Duration) error

	Open() error
	Close() error
	Ping() error
	Migrate() error
}
