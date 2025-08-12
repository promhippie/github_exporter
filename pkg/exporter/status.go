package exporter

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/v74/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// statusComponent represents a GitHub.com service shown on githubstatus.com.
type statusComponent string

const (
	compGitOperations statusComponent = "Git Operations"
	compWebhooks      statusComponent = "Webhooks"
	compAPIRequests   statusComponent = "API Requests"
	compIssues        statusComponent = "Issues"
	compPullRequests  statusComponent = "Pull Requests"
	compActions       statusComponent = "Actions"
	compPackages      statusComponent = "Packages"
	compPages         statusComponent = "Pages"
	compCodespaces    statusComponent = "Codespaces"
	compCopilot       statusComponent = "Copilot"
)

// statusComponents defines the ordered list of services we expose as gauges.
var statusComponents = []statusComponent{
	compGitOperations,
	compWebhooks,
	compAPIRequests,
	compIssues,
	compPullRequests,
	compActions,
	compPackages,
	compPages,
	compCodespaces,
	compCopilot,
}

func isStatusComponent(name string) bool {
	trimmed := strings.TrimSpace(name)
	for _, c := range statusComponents {
		if string(c) == trimmed {
			return true
		}
	}
	return false
}

// StatusCollector exposes gauges for GitHub component status.
type StatusCollector struct {
	client   *github.Client
	logger   *slog.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	GitOperationsUp *prometheus.Desc
	WebhooksUp      *prometheus.Desc
	APIRequestsUp   *prometheus.Desc
	IssuesUp        *prometheus.Desc
	PullRequestsUp  *prometheus.Desc
	ActionsUp       *prometheus.Desc
	PackagesUp      *prometheus.Desc
	PagesUp         *prometheus.Desc
	CodespacesUp    *prometheus.Desc
	CopilotUp       *prometheus.Desc
}

// NewStatusCollector returns a new StatusCollector with metric descriptors only.
func NewStatusCollector(logger *slog.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *StatusCollector {
	if failures != nil {
		failures.WithLabelValues("status").Add(0)
	}

	labels := []string{}
	return &StatusCollector{
		client:   client,
		logger:   logger.With("collector", "status"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		GitOperationsUp: prometheus.NewDesc(
			"github_status_git_operations_up",
			"Current health status of Git Operations on githubstatus.com",
			labels,
			nil,
		),
		WebhooksUp: prometheus.NewDesc(
			"github_status_webhooks_up",
			"Current health status of Webhooks on githubstatus.com",
			labels,
			nil,
		),
		APIRequestsUp: prometheus.NewDesc(
			"github_status_api_requests_up",
			"Current health status of API Requests on githubstatus.com",
			labels,
			nil,
		),
		IssuesUp: prometheus.NewDesc(
			"github_status_issues_up",
			"Current health status of Issues on githubstatus.com",
			labels,
			nil,
		),
		PullRequestsUp: prometheus.NewDesc(
			"github_status_pull_requests_up",
			"Current health status of Pull Requests on githubstatus.com",
			labels,
			nil,
		),
		ActionsUp: prometheus.NewDesc(
			"github_status_actions_up",
			"Current health status of Actions on githubstatus.com",
			labels,
			nil,
		),
		PackagesUp: prometheus.NewDesc(
			"github_status_packages_up",
			"Current health status of Packages on githubstatus.com",
			labels,
			nil,
		),
		PagesUp: prometheus.NewDesc(
			"github_status_pages_up",
			"Current health status of Pages on githubstatus.com",
			labels,
			nil,
		),
		CodespacesUp: prometheus.NewDesc(
			"github_status_codespaces_up",
			"Current health status of Codespaces on githubstatus.com",
			labels,
			nil,
		),
		CopilotUp: prometheus.NewDesc(
			"github_status_copilot_up",
			"Current health status of Copilot on githubstatus.com",
			labels,
			nil,
		),
	}
}

// Metrics returns descriptors for documentation generation.
func (c *StatusCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.GitOperationsUp,
		c.WebhooksUp,
		c.APIRequestsUp,
		c.IssuesUp,
		c.PullRequestsUp,
		c.ActionsUp,
		c.PackagesUp,
		c.PagesUp,
		c.CodespacesUp,
		c.CopilotUp,
	}
}

// Describe sends all possible descriptors.
func (c *StatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.GitOperationsUp
	ch <- c.WebhooksUp
	ch <- c.APIRequestsUp
	ch <- c.IssuesUp
	ch <- c.PullRequestsUp
	ch <- c.ActionsUp
	ch <- c.PackagesUp
	ch <- c.PagesUp
	ch <- c.CodespacesUp
	ch <- c.CopilotUp
}

func (c *StatusCollector) Collect(ch chan<- prometheus.Metric) {
	// Perform a single scrape of the status page and populate all gauges.
	// Follow redirects; treat "Operational" as up (1), everything else as down (0).
	client := &http.Client{Timeout: c.config.Timeout}

	now := time.Now()
	req, err := http.NewRequest("GET", "https://www.githubstatus.com/", nil)
	if err != nil {
		c.logger.Error("Failed to build status request", "err", err)
		if c.failures != nil {
			c.failures.WithLabelValues("status").Inc()
		}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Error("Failed to fetch status page", "err", err)
		if c.failures != nil {
			c.failures.WithLabelValues("status").Inc()
		}
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if c.duration != nil {
		c.duration.WithLabelValues("status").Observe(time.Since(now).Seconds())
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read status page", "err", err)
		if c.failures != nil {
			c.failures.WithLabelValues("status").Inc()
		}
		return
	}

	body := string(bodyBytes)
	statuses := extractStatusFromHTML(body)

	// Prepare the set of components to check → metric descriptors mapping.
	components := []struct {
		name statusComponent
		desc *prometheus.Desc
	}{
		{compGitOperations, c.GitOperationsUp},
		{compWebhooks, c.WebhooksUp},
		{compAPIRequests, c.APIRequestsUp},
		{compIssues, c.IssuesUp},
		{compPullRequests, c.PullRequestsUp},
		{compActions, c.ActionsUp},
		{compPackages, c.PackagesUp},
		{compPages, c.PagesUp},
		{compCodespaces, c.CodespacesUp},
		{compCopilot, c.CopilotUp},
	}

	// No labels for these metrics.
	labels := []string{}

	// Emit metrics for each component.
	for _, comp := range components {
		var up float64
		if ok, exists := statuses[string(comp.name)]; exists {
			if ok {
				up = 1.0
			} else {
				up = 0.0
			}
		} else {
			// If not found at all, consider as down and log for visibility.
			c.logger.Warn("Component status not found on status page", "component", comp.name)
			up = 0.0
		}
		c.logger.Debug("Component status scraped", "component", string(comp.name), "up", up)
		ch <- prometheus.MustNewConstMetric(
			comp.desc,
			prometheus.GaugeValue,
			up,
			labels...,
		)
	}
}

// extractStatusFromHTML parses the GitHub Status page HTML and returns a map of component name -> up (true if operational).
func extractStatusFromHTML(html string) map[string]bool {
	result := make(map[string]bool)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return result
	}

	doc.Find(".components-section .component-inner-container").Each(func(_ int, sel *goquery.Selection) {
		// Extract and normalize fields
		name := strings.TrimSpace(sel.Find(".name").Text())
		statusText := strings.ToLower(strings.TrimSpace(sel.Find(".component-status").Text()))

		if name == "" {
			return
		}

		// Only process names that are in our known list of components
		if !isStatusComponent(name) {
			return
		}

		// Determine up/down strictly from the visible status text
		if statusText == "operational" {
			result[name] = true
			return
		}

		// Any other value or missing status is down
		result[name] = false
	})

	return result
}
