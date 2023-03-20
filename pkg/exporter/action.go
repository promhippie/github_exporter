package exporter

import (
	"github.com/go-kit/log"
	"github.com/google/go-github/v50/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// ActionCollector collects metrics about the servers.
type ActionCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	WorkflowCount    *prometheus.Desc
	WorkflowDuration *prometheus.Desc
}

// NewActionCollector returns a new ActionCollector.
func NewActionCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *ActionCollector {
	if failures != nil {
		failures.WithLabelValues("action").Add(0)
	}

	labels := []string{"org", "repo", "event", "name", "job", "status", "head_branch", "runner_node_arch", "runner_node_os", "runner_node_type", "retry"}
	return &ActionCollector{
		client:   client,
		logger:   log.With(logger, "collector", "action"),
		failures: failures,
		duration: duration,
		config:   cfg,

		// github_runner_status
		// []string{"repo", "os", "name", "id", "busy"},
		// https://github.com/Spendesk/github-actions-exporter/blob/develop/pkg/metrics/get_runners_from_github.go

		// github_runner_organization_status
		// []string{"organization", "os", "name", "id", "busy"},
		// https://github.com/Spendesk/github-actions-exporter/blob/develop/pkg/metrics/get_runners_organization_from_github.go

		// github_runner_enterprise_status
		// []string{"os", "name", "id"}
		// https://github.com/Spendesk/github-actions-exporter/blob/develop/pkg/metrics/get_runners_enterprise_from_github.go

		WorkflowCount: prometheus.NewDesc(
			"github_action_workflow_count",
			"Number of workflow runs",
			labels,
			nil,
		),
		WorkflowDuration: prometheus.NewDesc(
			"github_action_workflow_duration_ms",
			"Duration of workflow runs",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *ActionCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.WorkflowCount,
		c.WorkflowDuration,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *ActionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.WorkflowCount
	ch <- c.WorkflowDuration
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ActionCollector) Collect(_ chan<- prometheus.Metric) {

}
