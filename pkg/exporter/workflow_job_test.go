package exporter

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

type StaticStore struct{}

func (s StaticStore) GetWorkflowJobRuns(owner, repo, workflow string) ([]*store.WorkflowRun, error) {
	_, _ = fmt.Fprintf(
		os.Stdout,
		"GetWorkflowJobRuns for %s/%s %s \n",
		owner,
		repo,
		workflow,
	)

	return nil, nil
}

func (s StaticStore) StoreWorkflowRunEvent(*github.WorkflowRunEvent) error {
	return nil
}

func (s StaticStore) GetWorkflowRuns(time.Duration) ([]*store.WorkflowRun, error) {
	return nil, nil
}

func (s StaticStore) PruneWorkflowRuns(time.Duration) error {
	return nil
}

func (s StaticStore) StoreWorkflowJobEvent(*github.WorkflowJobEvent) error {
	return nil
}

func (s StaticStore) GetWorkflowJobs(time.Duration) ([]*store.WorkflowJob, error) {
	return nil, nil
}

func (s StaticStore) PruneWorkflowJobs(time.Duration) error {
	return nil
}

func (s StaticStore) Open() (bool, error) {
	return true, nil
}

func (s StaticStore) Close() error {
	return nil
}

func (s StaticStore) Ping() (bool, error) {
	return true, nil
}

func (s StaticStore) Migrate() error {
	return nil
}

func TestWorkflowJobCollector(t *testing.T) {
	mockClient := &github.Client{}

	mockLogger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	mockStore := StaticStore{}

	mockFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_failures_total",
		Help: "Total number of test failures",
	}, []string{"type"})

	mockDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "test_duration_seconds",
		Help: "Duration of test",
	}, []string{"type"})

	mockConfig := config.Target{}

	collector := &WorkflowJobCollector{
		client:   mockClient,
		logger:   mockLogger,
		db:       mockStore,
		failures: mockFailures,
		duration: mockDuration,
		config:   mockConfig,
		Status: prometheus.NewDesc(
			"workflow_job_status",
			"Status of the workflow job",
			nil, nil,
		),
		Duration: prometheus.NewDesc(
			"workflow_job_duration_seconds",
			"Duration of the workflow job",
			nil, nil,
		),
		Creation: prometheus.NewDesc(
			"workflow_job_creation_timestamp_seconds",
			"Creation time of the workflow job",
			nil, nil,
		),
		Created: prometheus.NewDesc(
			"workflow_job_created_timestamp_seconds",
			"Created time of the workflow job",
			nil, nil,
		),
	}

	if collector.client != mockClient {
		t.Errorf("Expected client to be %v, got %v", mockClient, collector.client)
	}
	if collector.logger != mockLogger {
		t.Errorf("Expected logger to be %v, got %v", mockLogger, collector.logger)
	}
	if collector.db != mockStore {
		t.Errorf("Expected store to be %v, got %v", mockStore, collector.db)
	}
	if collector.failures != mockFailures {
		t.Errorf("Expected failures to be %v, got %v", mockFailures, collector.failures)
	}
	if collector.duration != mockDuration {
		t.Errorf("Expected duration to be %v, got %v", mockDuration, collector.duration)
	}
	if !reflect.DeepEqual(collector.config, mockConfig) {
		t.Errorf("Expected config to be %v, got %v", mockConfig, collector.config)
	}
}
