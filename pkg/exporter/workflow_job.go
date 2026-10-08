package exporter

import (
	"log/slog"
	"time"

	"github.com/google/go-github/v92/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// WorkflowJobCollector collects metrics about the servers.
type WorkflowJobCollector struct {
	client   *github.Client
	logger   *slog.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	Status             *prometheus.Desc
	Duration           *prometheus.Desc
	Creation           *prometheus.Desc
	Created            *prometheus.Desc
	Started            *prometheus.Desc
	CompletedTotal     *prometheus.Desc
	DurationTotal      *prometheus.Desc
	QueuedTotal        *prometheus.Desc
	QueueDurationTotal *prometheus.Desc
}

// NewWorkflowJobCollector returns a new WorkflowCollector.
func NewWorkflowJobCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *WorkflowJobCollector {
	if failures != nil {
		failures.WithLabelValues("action").Add(0)
	}

	labels := cfg.WorkflowJobs.Labels

	completionLabels := []string{
		"owner",
		"repo",
		"workflow_name",
		"name",
		"conclusion",
	}

	queueLabels := []string{
		"owner",
		"repo",
		"workflow_name",
		"runner_group_name",
		"conclusion",
	}

	return &WorkflowJobCollector{
		client:   client,
		logger:   logger.With("collector", "workflow_job"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		Status: prometheus.NewDesc(
			"github_workflow_job_status",
			"Status of workflow jobs",
			labels,
			nil,
		),
		Duration: prometheus.NewDesc(
			"github_workflow_job_duration_ms",
			"Duration of workflow runs",
			labels,
			nil,
		),
		Creation: prometheus.NewDesc(
			"github_workflow_job_duration_run_created_minutes",
			"Duration since the workflow run creation time in minutes",
			labels,
			nil,
		),
		Created: prometheus.NewDesc(
			"github_workflow_job_created_timestamp",
			"Timestamp when the workflow job have been created",
			labels,
			nil,
		),
		Started: prometheus.NewDesc(
			"github_workflow_job_started_timestamp",
			"Timestamp when the workflow job have been started",
			labels,
			nil,
		),
		CompletedTotal: prometheus.NewDesc(
			"github_workflow_job_completed_total",
			"Total number of completed workflow jobs",
			completionLabels,
			nil,
		),
		DurationTotal: prometheus.NewDesc(
			"github_workflow_job_duration_seconds_total",
			"Total duration of completed workflow jobs in seconds",
			completionLabels,
			nil,
		),
		QueuedTotal: prometheus.NewDesc(
			"github_workflow_job_queued_total",
			"Total number of completed workflow jobs that reported a queue time",
			queueLabels,
			nil,
		),
		QueueDurationTotal: prometheus.NewDesc(
			"github_workflow_job_queue_duration_seconds_total",
			"Total time workflow jobs spent queued before starting in seconds",
			queueLabels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *WorkflowJobCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.Status,
		c.Duration,
		c.Creation,
		c.Created,
		c.Started,
		c.CompletedTotal,
		c.DurationTotal,
		c.QueuedTotal,
		c.QueueDurationTotal,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *WorkflowJobCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
	ch <- c.Duration
	ch <- c.Creation
	ch <- c.Created
	ch <- c.Started
	ch <- c.CompletedTotal
	ch <- c.DurationTotal
	ch <- c.QueuedTotal
	ch <- c.QueueDurationTotal
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *WorkflowJobCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.db.PruneWorkflowJobs(
		c.config.WorkflowJobs.PurgeWindow,
	); err != nil {
		c.logger.Error("Failed to prune workflow jobs",
			"err", err,
		)
	}

	now := time.Now()
	records, err := c.db.GetWorkflowJobs(c.config.WorkflowJobs.Window)
	c.duration.WithLabelValues("workflow_job").Observe(time.Since(now).Seconds())

	if err != nil {
		c.logger.Error("Failed to fetch workflow jobs",
			"err", err,
		)

		c.failures.WithLabelValues("workflow_job").Inc()
		return
	}

	c.logger.Debug("Fetched workflow jobs",
		"count", len(records),
		"duration", time.Since(now),
	)

	for _, record := range records {
		c.logger.Debug("Collecting workflow job",
			"owner", record.Owner,
			"repo", record.Repo,
			"id", record.Identifier,
			"run_id", record.RunID,
		)

		labels := []string{}

		for _, label := range c.config.WorkflowJobs.Labels {
			labels = append(
				labels,
				record.ByLabel(label),
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Status,
			prometheus.GaugeValue,
			jobStatusToGauge(record.Status),
			labels...,
		)

		duration := float64(0)

		if record.CompletedAt > 0 && record.StartedAt > 0 {
			duration = float64((record.CompletedAt - record.StartedAt) * 1000)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Duration,
			prometheus.GaugeValue,
			duration,
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
			c.Started,
			prometheus.GaugeValue,
			float64(record.StartedAt),
			labels...,
		)
	}

	completions, err := c.db.GetWorkflowJobCompletions()

	if err != nil {
		c.logger.Error("Failed to fetch workflow job completions",
			"err", err,
		)

		c.failures.WithLabelValues("workflow_job").Inc()
		return
	}

	c.logger.Debug("Fetched workflow job completions",
		"count", len(completions),
	)

	for _, completion := range completions {
		labels := []string{
			completion.Owner,
			completion.Repo,
			completion.WorkflowName,
			completion.Name,
			completion.Conclusion,
		}

		ch <- prometheus.MustNewConstMetric(
			c.CompletedTotal,
			prometheus.CounterValue,
			float64(completion.Count),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DurationTotal,
			prometheus.CounterValue,
			completion.DurationSecondsTotal,
			labels...,
		)
	}

	queueDurations, err := c.db.GetWorkflowJobQueueDurations()

	if err != nil {
		c.logger.Error("Failed to fetch workflow job queue durations",
			"err", err,
		)

		c.failures.WithLabelValues("workflow_job").Inc()
		return
	}

	c.logger.Debug("Fetched workflow job queue durations",
		"count", len(queueDurations),
	)

	for _, queueDuration := range queueDurations {
		labels := []string{
			queueDuration.Owner,
			queueDuration.Repo,
			queueDuration.WorkflowName,
			queueDuration.RunnerGroupName,
			queueDuration.Conclusion,
		}

		ch <- prometheus.MustNewConstMetric(
			c.QueuedTotal,
			prometheus.CounterValue,
			float64(queueDuration.Count),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.QueueDurationTotal,
			prometheus.CounterValue,
			queueDuration.QueueDurationSecondsTotal,
			labels...,
		)
	}
}

func jobStatusToGauge(conclusion string) float64 {
	switch conclusion {
	case "queued":
		return 1.0
	case "waiting":
		return 2.0
	case "in_progress":
		return 3.0
	case "completed":
		return 4.0
	}

	return 0.0
}
