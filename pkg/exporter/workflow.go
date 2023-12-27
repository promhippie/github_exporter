package exporter

import (
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v57/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// WorkflowCollector collects metrics about the servers.
type WorkflowCollector struct {
	client   *github.Client
	logger   log.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	Status   *prometheus.Desc
	Duration *prometheus.Desc
	Creation *prometheus.Desc
	Created  *prometheus.Desc
	Updated  *prometheus.Desc
	Started  *prometheus.Desc
}

// NewWorkflowCollector returns a new WorkflowCollector.
func NewWorkflowCollector(logger log.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *WorkflowCollector {
	if failures != nil {
		failures.WithLabelValues("action").Add(0)
	}

	labels := cfg.Workflows.Labels.Value()
	return &WorkflowCollector{
		client:   client,
		logger:   log.With(logger, "collector", "workflow"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		Status: prometheus.NewDesc(
			"github_workflow_status",
			"Status of workflow runs",
			labels,
			nil,
		),
		Duration: prometheus.NewDesc(
			"github_workflow_duration_ms",
			"Duration of workflow runs",
			labels,
			nil,
		),
		Creation: prometheus.NewDesc(
			"github_workflow_duration_run_created_minutes",
			"Duration since the workflow run creation time in minutes",
			labels,
			nil,
		),
		Created: prometheus.NewDesc(
			"github_workflow_created_timestamp",
			"Timestammp when the workflow run have been created",
			labels,
			nil,
		),
		Updated: prometheus.NewDesc(
			"github_workflow_updated_timestamp",
			"Timestammp when the workflow run have been updated",
			labels,
			nil,
		),
		Started: prometheus.NewDesc(
			"github_workflow_started_timestamp",
			"Timestammp when the workflow run have been started",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *WorkflowCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.Status,
		c.Duration,
		c.Creation,
		c.Created,
		c.Updated,
		c.Started,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *WorkflowCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
	ch <- c.Duration
	ch <- c.Creation
	ch <- c.Created
	ch <- c.Updated
	ch <- c.Started
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *WorkflowCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.db.PruneWorkflowRuns(
		c.config.Workflows.Window,
	); err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to prune workflows",
			"err", err,
		)
	}

	now := time.Now()
	records, err := c.db.GetWorkflowRuns()
	c.duration.WithLabelValues("workflow").Observe(time.Since(now).Seconds())

	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to fetch workflows",
			"err", err,
		)

		c.failures.WithLabelValues("workflow").Inc()
		return
	}

	level.Debug(c.logger).Log(
		"msg", "Fetched workflows",
		"count", len(records),
		"duration", time.Since(now),
	)

	for _, record := range records {
		level.Debug(c.logger).Log(
			"msg", "Collecting workflow",
			"owner", record.Owner,
			"repo", record.Repo,
			"workflow_id", record.WorkflowID,
			"number", record.Number,
			"id", record.Identifier,
			"run_number", record.Number,
			"event", record.Event,
			"conclusion/status", record.Status,
			"created_at", record.CreatedAt,
			"updated_at", record.UpdatedAt,
			"run_started_at", record.StartedAt,
			"title", record.Title,
		)

		labels := []string{}

		for _, label := range c.config.Workflows.Labels.Value() {
			labels = append(
				labels,
				record.ByLabel(label),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Status,
			prometheus.GaugeValue,
			statusToGauge(record.Status),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Duration,
			prometheus.GaugeValue,
			float64((record.UpdatedAt-record.StartedAt)*1000),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Creation,
			prometheus.GaugeValue,
			time.Since(time.Unix(record.StartedAt, 0)).Minutes(),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Created,
			prometheus.GaugeValue,
			float64(record.CreatedAt),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Updated,
			prometheus.GaugeValue,
			float64(record.UpdatedAt),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Started,
			prometheus.GaugeValue,
			float64(record.StartedAt),
			labels...,
		)
	}
}

func statusToGauge(conclusion string) float64 {
	switch conclusion {
	case "completed":
		return 1.0
	case "action_required":
		return 2.0
	case "cancelled":
		return 3.0
	case "neutral":
		return 4.0
	case "skipped":
		return 5.0
	case "stale":
		return 6.0
	case "success":
		return 7.0
	case "timed_out":
		return 8.0
	case "in_progress":
		return 9.0
	case "queued":
		return 10.0
	case "requested":
		return 11.0
	case "waiting":
		return 12.0
	case "pending":
		return 13.0
	}

	return 0.0
}
