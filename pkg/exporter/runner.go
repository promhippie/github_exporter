package exporter

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v64/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
	"github.com/ryanuber/go-glob"
)

// RunnerCollector collects metrics about the runners.
type RunnerCollector struct {
	client   *github.Client
	logger   log.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	RepoOnline       *prometheus.Desc
	RepoBusy         *prometheus.Desc
	EnterpriseOnline *prometheus.Desc
	EnterpriseBusy   *prometheus.Desc
	OrgOnline        *prometheus.Desc
	OrgBusy          *prometheus.Desc
}

// NewRunnerCollector returns a new RunnerCollector.
func NewRunnerCollector(logger log.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *RunnerCollector {
	if failures != nil {
		failures.WithLabelValues("runner").Add(0)
	}

	labels := cfg.Runners.Labels.Value()
	return &RunnerCollector{
		client:   client,
		logger:   log.With(logger, "collector", "runner"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		RepoOnline: prometheus.NewDesc(
			"github_runner_repo_online",
			"Static metrics of runner is online or not",
			labels,
			nil,
		),
		RepoBusy: prometheus.NewDesc(
			"github_runner_repo_busy",
			"1 if the runner is busy, 0 otherwise",
			labels,
			nil,
		),
		EnterpriseOnline: prometheus.NewDesc(
			"github_runner_enterprise_online",
			"Static metrics of runner is online or not",
			labels,
			nil,
		),
		EnterpriseBusy: prometheus.NewDesc(
			"github_runner_enterprise_busy",
			"1 if the runner is busy, 0 otherwise",
			labels,
			nil,
		),
		OrgOnline: prometheus.NewDesc(
			"github_runner_org_online",
			"Static metrics of runner is online or not",
			labels,
			nil,
		),
		OrgBusy: prometheus.NewDesc(
			"github_runner_org_busy",
			"1 if the runner is busy, 0 otherwise",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *RunnerCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.RepoOnline,
		c.RepoBusy,
		c.EnterpriseOnline,
		c.EnterpriseBusy,
		c.OrgOnline,
		c.OrgBusy,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *RunnerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.RepoOnline
	ch <- c.RepoBusy
	ch <- c.EnterpriseOnline
	ch <- c.EnterpriseBusy
	ch <- c.OrgOnline
	ch <- c.OrgBusy
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *RunnerCollector) Collect(ch chan<- prometheus.Metric) {
	{
		collected := make([]string, 0)

		now := time.Now()
		records := c.repoRunners()
		c.duration.WithLabelValues("runner").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched repo runners",
			"count", len(records),
			"duration", time.Since(now),
		)

		for _, record := range records {
			if alreadyCollected(collected, record.GetName()) {
				level.Debug(c.logger).Log(
					"msg", "Already collected repo runner",
					"name", record.GetName(),
				)

				continue
			}

			collected = append(collected, record.GetName())

			var (
				online float64
			)

			level.Debug(c.logger).Log(
				"msg", "Collecting repo runner",
				"name", record.GetName(),
			)

			labels := []string{}

			for _, label := range c.config.Runners.Labels.Value() {
				labels = append(
					labels,
					record.ByLabel(label),
				)
			}

			if record.GetStatus() == "online" {
				online = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.RepoOnline,
				prometheus.GaugeValue,
				online,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.RepoBusy,
				prometheus.GaugeValue,
				boolToFloat64(*record.Busy),
				labels...,
			)
		}
	}

	{
		collected := make([]string, 0)

		now := time.Now()
		records := c.enterpriseRunners()
		c.duration.WithLabelValues("runner").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched enterprise runners",
			"count", len(records),
			"duration", time.Since(now),
		)

		for _, record := range records {
			if alreadyCollected(collected, record.GetName()) {
				level.Debug(c.logger).Log(
					"msg", "Already collected enterprise runner",
					"name", record.GetName(),
				)

				continue
			}

			collected = append(collected, record.GetName())

			var (
				online float64
			)

			level.Debug(c.logger).Log(
				"msg", "Collecting enterprise runner",
				"name", record.GetName(),
			)

			labels := []string{}

			for _, label := range c.config.Runners.Labels.Value() {
				labels = append(
					labels,
					record.ByLabel(label),
				)
			}

			if record.GetStatus() == "online" {
				online = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.EnterpriseOnline,
				prometheus.GaugeValue,
				online,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.EnterpriseBusy,
				prometheus.GaugeValue,
				boolToFloat64(*record.Busy),
				labels...,
			)
		}
	}

	{
		collected := make([]string, 0)

		now := time.Now()
		records := c.orgRunners()
		c.duration.WithLabelValues("runner").Observe(time.Since(now).Seconds())

		level.Debug(c.logger).Log(
			"msg", "Fetched org runners",
			"count", len(records),
			"duration", time.Since(now),
		)

		for _, record := range records {
			if alreadyCollected(collected, record.GetName()) {
				level.Debug(c.logger).Log(
					"msg", "Already collected org runner",
					"name", record.GetName(),
				)

				continue
			}

			collected = append(collected, record.GetName())

			var (
				online float64
			)

			level.Debug(c.logger).Log(
				"msg", "Collecting org runner",
				"name", record.GetName(),
			)

			labels := []string{}

			for _, label := range c.config.Runners.Labels.Value() {
				labels = append(
					labels,
					record.ByLabel(label),
				)
			}

			if record.GetStatus() == "online" {
				online = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.OrgOnline,
				prometheus.GaugeValue,
				online,
				labels...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.OrgBusy,
				prometheus.GaugeValue,
				boolToFloat64(*record.Busy),
				labels...,
			)
		}
	}
}

func (c *RunnerCollector) repoRunners() []runner {
	collected := make([]string, 0)
	result := make([]runner, 0)

	for _, name := range c.config.Repos.Value() {
		n := strings.Split(name, "/")

		if len(n) != 2 {
			level.Error(c.logger).Log(
				"msg", "Invalid repo name",
				"name", name,
			)

			c.failures.WithLabelValues("runner").Inc()
			continue
		}

		splitOwner, splitName := n[0], n[1]

		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		repos, err := reposByOwnerAndName(ctx, c.client, splitOwner, splitName, c.config.PerPage)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch repos",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("runner").Inc()
			continue
		}

		level.Debug(c.logger).Log(
			"msg", "Fetched repos for runners",
			"count", len(repos),
		)

		for _, repo := range repos {
			if !glob.Glob(name, *repo.FullName) {
				continue
			}

			if alreadyCollected(collected, repo.GetFullName()) {
				level.Debug(c.logger).Log(
					"msg", "Already collected repo",
					"name", repo.GetFullName(),
				)

				continue
			}

			collected = append(collected, repo.GetFullName())

			records, err := c.pagedRepoRunners(ctx, *repo.Owner.Login, *repo.Name)

			if err != nil {
				level.Error(c.logger).Log(
					"msg", "Failed to fetch repo runners",
					"name", name,
					"err", err,
				)

				c.failures.WithLabelValues("runner").Inc()
				continue
			}

			for _, row := range records {
				result = append(result, runner{
					Owner:  name,
					Runner: row,
				})
			}
		}
	}

	return result
}

func (c *RunnerCollector) pagedRepoRunners(ctx context.Context, owner, name string) ([]*github.Runner, error) {
	opts := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{
			PerPage: c.config.PerPage,
		},
	}

	var (
		runners []*github.Runner
	)

	for {
		result, resp, err := c.client.Actions.ListRunners(
			ctx,
			owner,
			name,
			opts,
		)

		if err != nil {
			closeBody(resp)
			return nil, err
		}

		runners = append(
			runners,
			result.Runners...,
		)

		if resp.NextPage == 0 {
			closeBody(resp)
			break
		}

		closeBody(resp)
		opts.Page = resp.NextPage
	}

	return runners, nil
}

func (c *RunnerCollector) enterpriseRunners() []runner {
	result := make([]runner, 0)

	for _, name := range c.config.Enterprises.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		records, err := c.pagedEnterpriseRunners(ctx, name)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch enterprise runners",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("runner").Inc()
			continue
		}

		for _, row := range records {
			result = append(result, runner{
				Owner:  name,
				Runner: row,
			})
		}
	}

	return result
}

func (c *RunnerCollector) pagedEnterpriseRunners(ctx context.Context, name string) ([]*github.Runner, error) {
	opts := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{
			PerPage: c.config.PerPage,
		},
	}

	var (
		runners []*github.Runner
	)

	for {
		result, resp, err := c.client.Enterprise.ListRunners(
			ctx,
			name,
			opts,
		)

		if err != nil {
			closeBody(resp)
			return nil, err
		}

		runners = append(
			runners,
			result.Runners...,
		)

		if resp.NextPage == 0 {
			closeBody(resp)
			break
		}

		closeBody(resp)
		opts.Page = resp.NextPage
	}

	return runners, nil
}

func (c *RunnerCollector) orgRunners() []runner {
	result := make([]runner, 0)

	for _, name := range c.config.Orgs.Value() {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		records, err := c.pagedOrgRunners(ctx, name)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch org runners",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("runner").Inc()
			continue
		}

		for _, row := range records {
			result = append(result, runner{
				Owner:  name,
				Runner: row,
			})
		}
	}

	return result
}

func (c *RunnerCollector) pagedOrgRunners(ctx context.Context, name string) ([]*github.Runner, error) {
	opts := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{
			PerPage: c.config.PerPage,
		},
	}

	var (
		runners []*github.Runner
	)

	for {
		result, resp, err := c.client.Actions.ListOrganizationRunners(
			ctx,
			name,
			opts,
		)

		if err != nil {
			closeBody(resp)
			return nil, err
		}

		runners = append(
			runners,
			result.Runners...,
		)

		if resp.NextPage == 0 {
			closeBody(resp)
			break
		}

		closeBody(resp)
		opts.Page = resp.NextPage
	}

	return runners, nil
}

type runner struct {
	Owner string
	*github.Runner
}

// AggregateLabels Aggregate custom labels into comma delimited string
func (r *runner) AggregateLabels() string {
	var aggLabels []string
	for _, label := range r.Labels {
		if label != nil && label.Type != nil {
			aggLabels = append(aggLabels, *label.Name)
		}
	}

	sort.Strings(aggLabels)
	return strings.Join(aggLabels, ",")
}

func (r *runner) ByLabel(label string) string {
	switch label {
	case "owner":
		return r.Owner
	case "id":
		return strconv.FormatInt(r.GetID(), 10)
	case "name":
		return r.GetName()
	case "os":
		return r.GetOS()
	case "status":
		return r.GetStatus()
	case "labels":
		return r.AggregateLabels()
	}

	return ""
}
