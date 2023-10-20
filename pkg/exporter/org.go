package exporter

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v56/github"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// OrgCollector collects metrics about the servers.
type OrgCollector struct {
	client   *github.Client
	logger   log.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	PublicRepos       *prometheus.Desc
	PublicGists       *prometheus.Desc
	PrivateGists      *prometheus.Desc
	Followers         *prometheus.Desc
	Following         *prometheus.Desc
	Collaborators     *prometheus.Desc
	DiskUsage         *prometheus.Desc
	PrivateReposTotal *prometheus.Desc
	PrivateReposOwned *prometheus.Desc
	Seats             *prometheus.Desc
	FilledSeats       *prometheus.Desc
	Created           *prometheus.Desc
	Updated           *prometheus.Desc
}

// NewOrgCollector returns a new OrgCollector.
func NewOrgCollector(logger log.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *OrgCollector {
	if failures != nil {
		failures.WithLabelValues("org").Add(0)
	}

	labels := []string{"name"}
	return &OrgCollector{
		client:   client,
		logger:   log.With(logger, "collector", "org"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		PublicRepos: prometheus.NewDesc(
			"github_org_public_repos",
			"Number of public repositories from org",
			labels,
			nil,
		),
		PublicGists: prometheus.NewDesc(
			"github_org_public_gists",
			"Number of public gists from org",
			labels,
			nil,
		),
		PrivateGists: prometheus.NewDesc(
			"github_org_private_gists",
			"Number of private gists from org",
			labels,
			nil,
		),
		Followers: prometheus.NewDesc(
			"github_org_followers",
			"Number of followers for org",
			labels,
			nil,
		),
		Following: prometheus.NewDesc(
			"github_org_following",
			"Number of following other users by org",
			labels,
			nil,
		),
		Collaborators: prometheus.NewDesc(
			"github_org_collaborators",
			"Number of collaborators within org",
			labels,
			nil,
		),
		DiskUsage: prometheus.NewDesc(
			"github_org_disk_usage",
			"Used diskspace by the org",
			labels,
			nil,
		),
		PrivateReposTotal: prometheus.NewDesc(
			"github_org_private_repos_total",
			"Total amount of private repositories",
			labels,
			nil,
		),
		PrivateReposOwned: prometheus.NewDesc(
			"github_org_private_repos_owned",
			"Owned private repositories by org",
			labels,
			nil,
		),
		FilledSeats: prometheus.NewDesc(
			"github_org_filled_seats",
			"Filled seats for org",
			labels,
			nil,
		),
		Seats: prometheus.NewDesc(
			"github_org_seats",
			"Seats for org",
			labels,
			nil,
		),
		Created: prometheus.NewDesc(
			"github_org_create_timestamp",
			"Timestamp of the creation of org",
			labels,
			nil,
		),
		Updated: prometheus.NewDesc(
			"github_org_updated_timestamp",
			"Timestamp of the last modification of org",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *OrgCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.PublicRepos,
		c.PublicGists,
		c.PrivateGists,
		c.Followers,
		c.Following,
		c.Collaborators,
		c.DiskUsage,
		c.PrivateReposTotal,
		c.PrivateReposOwned,
		c.FilledSeats,
		c.Seats,
		c.Created,
		c.Updated,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *OrgCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PublicRepos
	ch <- c.PublicGists
	ch <- c.PrivateGists
	ch <- c.Followers
	ch <- c.Following
	ch <- c.Collaborators
	ch <- c.DiskUsage
	ch <- c.PrivateReposTotal
	ch <- c.PrivateReposOwned
	ch <- c.Seats
	ch <- c.FilledSeats
	ch <- c.Created
	ch <- c.Updated
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *OrgCollector) Collect(ch chan<- prometheus.Metric) {
	collected := make([]string, 0)

	for _, name := range c.config.Orgs.Value() {
		if alreadyCollected(collected, name) {
			level.Debug(c.logger).Log(
				"msg", "Already collected org",
				"name", name,
			)

			continue
		}

		collected = append(collected, name)

		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		now := time.Now()
		record, resp, err := c.client.Organizations.Get(ctx, name)
		c.duration.WithLabelValues("org").Observe(time.Since(now).Seconds())
		defer closeBody(resp)

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("org").Inc()
			continue
		}

		level.Debug(c.logger).Log(
			"msg", "Collecting org",
			"name", name,
		)

		labels := []string{
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.PublicRepos,
			prometheus.GaugeValue,
			float64(record.GetPublicRepos()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PublicGists,
			prometheus.GaugeValue,
			float64(record.GetPublicGists()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateGists,
			prometheus.GaugeValue,
			float64(record.GetPrivateGists()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Followers,
			prometheus.GaugeValue,
			float64(record.GetFollowers()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Following,
			prometheus.GaugeValue,
			float64(record.GetFollowing()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Collaborators,
			prometheus.GaugeValue,
			float64(record.GetCollaborators()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DiskUsage,
			prometheus.GaugeValue,
			float64(record.GetDiskUsage()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateReposTotal,
			prometheus.GaugeValue,
			float64(record.GetTotalPrivateRepos()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PrivateReposOwned,
			prometheus.GaugeValue,
			float64(record.GetOwnedPrivateRepos()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Seats,
			prometheus.GaugeValue,
			float64(record.GetPlan().GetSeats()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilledSeats,
			prometheus.GaugeValue,
			float64(record.GetPlan().GetFilledSeats()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Created,
			prometheus.GaugeValue,
			float64(record.GetCreatedAt().Unix()),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Updated,
			prometheus.GaugeValue,
			float64(record.GetUpdatedAt().Unix()),
			labels...,
		)
	}
}
