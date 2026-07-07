package exporter

import (
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/v89/github"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

type StaticStore struct{}

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

func (s StaticStore) GetWorkflowJobCompletions() ([]*store.WorkflowJobCompletionAggregate, error) {
	return nil, nil
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
		Started: prometheus.NewDesc(
			"workflow_job_started_timestamp_seconds",
			"Started time of the workflow job",
			nil, nil,
		),
		CompletedTotal: prometheus.NewDesc(
			"workflow_job_completed_total",
			"Total number of completed workflow jobs",
			nil, nil,
		),
		DurationSecondsTotal: prometheus.NewDesc(
			"workflow_job_duration_seconds_total",
			"Total duration of completed workflow jobs in seconds",
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

type completionStore struct {
	StaticStore
	completions []*store.WorkflowJobCompletionAggregate
}

func (s completionStore) GetWorkflowJobCompletions() ([]*store.WorkflowJobCompletionAggregate, error) {
	return s.completions, nil
}

func TestWorkflowJobCollectorCounters(t *testing.T) {
	mockLogger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	mockFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_failures_total",
		Help: "Total number of test failures",
	}, []string{"type"})

	mockDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "test_duration_seconds",
		Help: "Duration of test",
	}, []string{"type"})

	completions := []*store.WorkflowJobCompletionAggregate{
		{
			Owner:                "promhippie",
			Repo:                 "github_exporter",
			WorkflowName:         "CI",
			Name:                 "test",
			Conclusion:           "success",
			Count:                2,
			DurationSecondsTotal: 42.5,
		},
		{
			Owner:                "promhippie",
			Repo:                 "github_exporter",
			WorkflowName:         "CI",
			Name:                 "test",
			Conclusion:           "failure",
			Count:                1,
			DurationSecondsTotal: 10.0,
		},
	}

	store := completionStore{completions: completions}
	collector := NewWorkflowJobCollector(
		mockLogger,
		nil,
		store,
		mockFailures,
		mockDuration,
		config.Target{},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expected := map[string]float64{
		"github_workflow_job_completed_total":        3,
		"github_workflow_job_duration_seconds_total": 52.5,
	}

	for name, expectedValue := range expected {
		value := metricFamilyValue(t, metrics, name)
		if value != expectedValue {
			t.Errorf("expected %s to be %v, got %v", name, expectedValue, value)
		}
	}
}

func metricFamilyValue(t *testing.T, metrics []*dto.MetricFamily, name string) float64 {
	t.Helper()

	for _, mf := range metrics {
		if mf.GetName() != name {
			continue
		}

		var total float64

		for _, m := range mf.GetMetric() {
			if m.Counter != nil {
				total += m.Counter.GetValue()
			}
		}

		return total
	}

	t.Errorf("metric family %s not found", name)
	return 0
}
