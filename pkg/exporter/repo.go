package exporter

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v35/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// RepoCollector collects metrics about the servers.
type RepoCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	Forked           *prometheus.Desc
	Forks            *prometheus.Desc
	Network          *prometheus.Desc
	Issues           *prometheus.Desc
	Stargazers       *prometheus.Desc
	Subscribers      *prometheus.Desc
	Watchers         *prometheus.Desc
	Size             *prometheus.Desc
	AllowRebaseMerge *prometheus.Desc
	AllowSquashMerge *prometheus.Desc
	AllowMergeCommit *prometheus.Desc
	Archived         *prometheus.Desc
	Private          *prometheus.Desc
	HasIssues        *prometheus.Desc
	HasWiki          *prometheus.Desc
	HasPages         *prometheus.Desc
	HasProjects      *prometheus.Desc
	HasDownloads     *prometheus.Desc
	Pushed           *prometheus.Desc
	Created          *prometheus.Desc
	Updated          *prometheus.Desc
}

// NewRepoCollector returns a new RepoCollector.
func NewRepoCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *RepoCollector {
	if failures != nil {
		failures.WithLabelValues("repo").Add(0)
	}

	labels := []string{"owner", "name"}
	return &RepoCollector{
		client:   client,
		logger:   log.With(logger, "collector", "repo"),
		failures: failures,
		duration: duration,
		config:   cfg,

		Pushed: prometheus.NewDesc(
			"github_repo_pushed_timestamp",
			"Timestamp of the last push to repo",
			labels,
			nil,
		),
		Created: prometheus.NewDesc(
			"github_repo_created_timestamp",
			"Timestamp of the creation of repo",
			labels,
			nil,
		),
		Updated: prometheus.NewDesc(
			"github_repo_updated_timestamp",
			"Timestamp of the last modification of repo",
			labels,
			nil,
		),
		Forked: prometheus.NewDesc(
			"github_repo_forked",
			"Show if this repository is a forked repository",
			labels,
			nil,
		),
		Forks: prometheus.NewDesc(
			"github_repo_forks",
			"How often has this repository been forked",
			labels,
			nil,
		),
		Network: prometheus.NewDesc(
			"github_repo_network",
			"Number of repositories in the network",
			labels,
			nil,
		),
		Issues: prometheus.NewDesc(
			"github_repo_issues",
			"Number of open issues on this repository",
			labels,
			nil,
		),
		Stargazers: prometheus.NewDesc(
			"github_repo_stargazers",
			"Number of stargazers on this repository",
			labels,
			nil,
		),
		Subscribers: prometheus.NewDesc(
			"github_repo_subscribers",
			"Number of subscribers on this repository",
			labels,
			nil,
		),
		Watchers: prometheus.NewDesc(
			"github_repo_watchers",
			"Number of watchers on this repository",
			labels,
			nil,
		),
		Size: prometheus.NewDesc(
			"github_repo_size",
			"Size of the repository content",
			labels,
			nil,
		),
		AllowRebaseMerge: prometheus.NewDesc(
			"github_repo_allow_rebase_merge",
			"Show if this repository allows rebase merges",
			labels,
			nil,
		),
		AllowSquashMerge: prometheus.NewDesc(
			"github_repo_allow_squash_merge",
			"Show if this repository allows squash merges",
			labels,
			nil,
		),
		AllowMergeCommit: prometheus.NewDesc(
			"github_repo_allow_merge_commit",
			"Show if this repository allows merge commits",
			labels,
			nil,
		),
		Archived: prometheus.NewDesc(
			"github_repo_archived",
			"Show if this repository have been archived",
			labels,
			nil,
		),
		Private: prometheus.NewDesc(
			"github_repo_private",
			"Show iof this repository is private",
			labels,
			nil,
		),
		HasIssues: prometheus.NewDesc(
			"github_repo_has_issues",
			"Show if this repository got issues enabled",
			labels,
			nil,
		),
		HasWiki: prometheus.NewDesc(
			"github_repo_has_wiki",
			"Show if this repository got wiki enabled",
			labels,
			nil,
		),
		HasPages: prometheus.NewDesc(
			"github_repo_has_pages",
			"Show if this repository got pages enabled",
			labels,
			nil,
		),
		HasProjects: prometheus.NewDesc(
			"github_repo_has_projects",
			"Show if this repository got projects enabled",
			labels,
			nil,
		),
		HasDownloads: prometheus.NewDesc(
			"github_repo_has_downloads",
			"Show if this repository got downloads enabled",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *RepoCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.Forked,
		c.Forks,
		c.Network,
		c.Issues,
		c.Stargazers,
		c.Subscribers,
		c.Watchers,
		c.Size,
		c.AllowRebaseMerge,
		c.AllowSquashMerge,
		c.AllowMergeCommit,
		c.Archived,
		c.Private,
		c.HasIssues,
		c.HasWiki,
		c.HasPages,
		c.HasProjects,
		c.HasDownloads,
		c.Pushed,
		c.Created,
		c.Updated,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *RepoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Forked
	ch <- c.Forks
	ch <- c.Network
	ch <- c.Issues
	ch <- c.Stargazers
	ch <- c.Subscribers
	ch <- c.Watchers
	ch <- c.Size
	ch <- c.AllowRebaseMerge
	ch <- c.AllowSquashMerge
	ch <- c.AllowMergeCommit
	ch <- c.Archived
	ch <- c.Private
	ch <- c.HasIssues
	ch <- c.HasWiki
	ch <- c.HasPages
	ch <- c.HasProjects
	ch <- c.HasDownloads
	ch <- c.Pushed
	ch <- c.Created
	ch <- c.Updated
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *RepoCollector) Collect(ch chan<- prometheus.Metric) {
	for _, name := range c.config.Repos.Value() {
		n := strings.Split(name, "/")

		if len(n) != 2 {
			level.Error(c.logger).Log(
				"msg", "Invalid repo name",
				"name", name,
			)

			c.failures.WithLabelValues("repo").Inc()
			continue
		}

		owner, repo := n[0], n[1]

		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		now := time.Now()
		record, _, err := c.client.Repositories.Get(ctx, owner, repo)
		c.duration.WithLabelValues("repo").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch repo",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("repo").Inc()
			continue
		}

		labels := []string{
			owner,
			repo,
		}

		if record.Fork != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Forked,
				prometheus.GaugeValue,
				boolToFloat64(*record.Fork),
				labels...,
			)
		}

		if record.ForksCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Forks,
				prometheus.GaugeValue,
				float64(*record.ForksCount),
				labels...,
			)
		}

		if record.NetworkCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Network,
				prometheus.GaugeValue,
				float64(*record.NetworkCount),
				labels...,
			)
		}

		if record.OpenIssuesCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Issues,
				prometheus.GaugeValue,
				float64(*record.OpenIssuesCount),
				labels...,
			)
		}

		if record.StargazersCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Stargazers,
				prometheus.GaugeValue,
				float64(*record.StargazersCount),
				labels...,
			)
		}

		if record.SubscribersCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Subscribers,
				prometheus.GaugeValue,
				float64(*record.SubscribersCount),
				labels...,
			)
		}

		if record.WatchersCount != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Watchers,
				prometheus.GaugeValue,
				float64(*record.WatchersCount),
				labels...,
			)
		}

		if record.Size != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Size,
				prometheus.GaugeValue,
				float64(*record.Size),
				labels...,
			)
		}

		if record.AllowRebaseMerge != nil {
			ch <- prometheus.MustNewConstMetric(
				c.AllowRebaseMerge,
				prometheus.GaugeValue,
				boolToFloat64(*record.AllowRebaseMerge),
				labels...,
			)
		}

		if record.AllowSquashMerge != nil {
			ch <- prometheus.MustNewConstMetric(
				c.AllowSquashMerge,
				prometheus.GaugeValue,
				boolToFloat64(*record.AllowSquashMerge),
				labels...,
			)
		}

		if record.AllowMergeCommit != nil {
			ch <- prometheus.MustNewConstMetric(
				c.AllowMergeCommit,
				prometheus.GaugeValue,
				boolToFloat64(*record.AllowMergeCommit),
				labels...,
			)
		}

		if record.Archived != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Archived,
				prometheus.GaugeValue,
				boolToFloat64(*record.Archived),
				labels...,
			)
		}

		if record.Private != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Private,
				prometheus.GaugeValue,
				boolToFloat64(*record.Private),
				labels...,
			)
		}

		if record.HasIssues != nil {
			ch <- prometheus.MustNewConstMetric(
				c.HasIssues,
				prometheus.GaugeValue,
				boolToFloat64(*record.HasIssues),
				labels...,
			)
		}

		if record.HasWiki != nil {
			ch <- prometheus.MustNewConstMetric(
				c.HasWiki,
				prometheus.GaugeValue,
				boolToFloat64(*record.HasWiki),
				labels...,
			)
		}

		if record.HasPages != nil {
			ch <- prometheus.MustNewConstMetric(
				c.HasPages,
				prometheus.GaugeValue,
				boolToFloat64(*record.HasPages),
				labels...,
			)
		}

		if record.HasProjects != nil {
			ch <- prometheus.MustNewConstMetric(
				c.HasProjects,
				prometheus.GaugeValue,
				boolToFloat64(*record.HasProjects),
				labels...,
			)
		}

		if record.HasDownloads != nil {
			ch <- prometheus.MustNewConstMetric(
				c.HasDownloads,
				prometheus.GaugeValue,
				boolToFloat64(*record.HasDownloads),
				labels...,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Pushed,
			prometheus.GaugeValue,
			float64(record.PushedAt.Unix()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Created,
			prometheus.GaugeValue,
			float64(record.CreatedAt.Unix()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Updated,
			prometheus.GaugeValue,
			float64(record.UpdatedAt.Unix()),
			labels...,
		)
	}
}

func boolToFloat64(val bool) float64 {
	if val {
		return 1.0
	}

	return 0.0
}
