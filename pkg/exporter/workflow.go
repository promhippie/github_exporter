package exporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v50/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/ryanuber/go-glob"
)

// WorkflowCollector collects metrics about the servers.
type WorkflowCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	Status   *prometheus.Desc
	Duration *prometheus.Desc
}

// NewWorkflowCollector returns a new WorkflowCollector.
func NewWorkflowCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *WorkflowCollector {
	if failures != nil {
		failures.WithLabelValues("action").Add(0)
	}

	labels := []string{"owner", "repo", "event", "name", "status", "head_branch", "run", "retry"}
	return &WorkflowCollector{
		client:   client,
		logger:   log.With(logger, "collector", "workflow"),
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
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *WorkflowCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.Status,
		c.Duration,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *WorkflowCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
	ch <- c.Duration
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *WorkflowCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	records := c.repoWorkflows()
	c.duration.WithLabelValues("runner").Observe(time.Since(now).Seconds())

	for _, record := range records {
		labels := []string{
			record.GetRepository().GetOwner().GetLogin(),
			record.GetRepository().GetName(),
			record.GetEvent(),
			record.GetName(),
			record.GetStatus(),
			record.GetHeadBranch(),
			strconv.Itoa(record.GetRunNumber()),
			strconv.Itoa(record.GetRunAttempt()),
		}

		ch <- prometheus.MustNewConstMetric(
			c.Status,
			prometheus.GaugeValue,
			conclusionToGauge(record.GetConclusion()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Duration,
			prometheus.GaugeValue,
			float64((record.GetUpdatedAt().Time.Unix()-record.GetCreatedAt().Time.Unix())*1000),
			labels...,
		)
	}
}

func (c *WorkflowCollector) repoWorkflows() []*github.WorkflowRun {
	result := make([]*github.WorkflowRun, 0)

	for _, name := range c.config.Repos.Value() {
		n := strings.Split(name, "/")

		if len(n) != 2 {
			level.Error(c.logger).Log(
				"msg", "Invalid repo name",
				"name", name,
			)

			c.failures.WithLabelValues("workflow").Inc()
			continue
		}

		splitOwner, splitName := n[0], n[1]

		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		repos, err := reposByOwnerAndName(ctx, c.client, splitOwner, splitName)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch repos",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("workflow").Inc()
			continue
		}

		for _, repo := range repos {
			if !glob.Glob(name, *repo.FullName) {
				continue
			}

			records, err := c.pagedRepoWorkflows(ctx, *repo.Owner.Login, *repo.Name)

			if err != nil {
				level.Error(c.logger).Log(
					"msg", "Failed to fetch repo workflows",
					"name", name,
					"err", err,
				)

				c.failures.WithLabelValues("workflow").Inc()
				continue
			}

			result = append(result, records...)
		}
	}

	return result
}

func (c *WorkflowCollector) pagedRepoWorkflows(ctx context.Context, owner, name string) ([]*github.WorkflowRun, error) {
	startWindow := time.Now().Add(
		time.Duration(-12) * time.Hour,
	).Format(time.RFC3339)

	opts := &github.ListWorkflowRunsOptions{
		Created: fmt.Sprintf(">=%s", startWindow),
		ListOptions: github.ListOptions{
			PerPage: 200,
		},
	}

	var (
		workflows []*github.WorkflowRun
	)

	for {
		result, resp, err := c.client.Actions.ListRepositoryWorkflowRuns(
			ctx,
			owner,
			name,
			opts,
		)

		if err != nil {
			return nil, err
		}

		workflows = append(
			workflows,
			result.WorkflowRuns...,
		)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return workflows, nil
}

func conclusionToGauge(conclusion string) float64 {
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
