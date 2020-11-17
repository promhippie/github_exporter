package exporter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v32/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// OrgCollector collects metrics about organizations
type OrgCollector struct {
	client   *github.Client
	logger   log.Logger
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
	Created           *prometheus.Desc
	Updated           *prometheus.Desc
}

// NewOrgCollector returns a new OrgCollector.
func NewOrgCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *OrgCollector {
	failures.WithLabelValues("org").Add(0)

	labels := []string{"name"}
	return &OrgCollector{
		client:   client,
		logger:   log.With(logger, "collector", "org"),
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
	ch <- c.Created
	ch <- c.Updated
}

// Consume the full list of orgs from the API page by page
func fetchAllOrgs(c *OrgCollector) []string {
	var orgNames []string

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	var lastSeen int64 = 0
	var errorCount int = 0

	for {
		var params *github.OrganizationsListOptions

		// paging is done via the API ?since parameter
		if lastSeen > 0 {
			params = &github.OrganizationsListOptions{Since: lastSeen}
		} else {
			params = nil
		}

		orgs, _, err := c.client.Organizations.ListAll(ctx, params)
		if err != nil {
			errorCount++

			level.Error(c.logger).Log(
				"msg", fmt.Sprintf("Failed to list orgs %d", errorCount),
				"err", err,
			)
			c.failures.WithLabelValues("org").Inc()

			if errorCount >= 3 {
				break
			}
			continue
		}

		if len(orgs) == 0 {
			break
		}

		// extract array of string Organization.Login
		names := make([]string, len(orgs))
		for i := 0; i < len(orgs); i++ {
			names[i] = *orgs[i].Login
		}
		orgNames = append(orgNames, names...)

		// track last seen Organization.ID for paging
		lastSeen = *orgs[len(orgs)-1].ID
	}

	return orgNames
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *OrgCollector) Collect(ch chan<- prometheus.Metric) {
	var orgNames []string
	var err error

	// fetch ALL github orgs, or just the few supplied in config.Orgs
	if c.config.Orgs.Value()[0] == "*" {
		orgNames = fetchAllOrgs(c)
	} else {
		orgNames = c.config.Orgs.Value()
	}

	if len(orgNames) == 0 {
		level.Error(c.logger).Log(
			"msg", "Aborted scrape as failed to fetch list of orgs!",
			"err", err,
		)
		c.failures.WithLabelValues("org").Inc()
		return
	}

	for _, name := range orgNames {
		ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
		defer cancel()

		now := time.Now()
		record, _, err := c.client.Organizations.Get(ctx, name)
		c.duration.WithLabelValues("org").Observe(time.Since(now).Seconds())

		if err != nil {
			level.Error(c.logger).Log(
				"msg", "Failed to fetch org",
				"name", name,
				"err", err,
			)

			c.failures.WithLabelValues("org").Inc()
			continue
		}

		level.Info(c.logger).Log("org", name)
		labels := []string{
			name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.PublicRepos,
			prometheus.GaugeValue,
			float64(*record.PublicRepos),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PublicGists,
			prometheus.GaugeValue,
			float64(*record.PublicGists),
			labels...,
		)

		if record.PrivateGists != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateGists,
				prometheus.GaugeValue,
				float64(*record.PrivateGists),
				labels...,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.Followers,
			prometheus.GaugeValue,
			float64(*record.Followers),
			labels...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Following,
			prometheus.GaugeValue,
			float64(*record.Following),
			labels...,
		)

		if record.Collaborators != nil {
			ch <- prometheus.MustNewConstMetric(
				c.Collaborators,
				prometheus.GaugeValue,
				float64(*record.Collaborators),
				labels...,
			)
		}

		if record.DiskUsage != nil {
			ch <- prometheus.MustNewConstMetric(
				c.DiskUsage,
				prometheus.GaugeValue,
				float64(*record.DiskUsage),
				labels...,
			)
		}

		if record.TotalPrivateRepos != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateReposTotal,
				prometheus.GaugeValue,
				float64(*record.TotalPrivateRepos),
				labels...,
			)
		}

		if record.OwnedPrivateRepos != nil {
			ch <- prometheus.MustNewConstMetric(
				c.PrivateReposOwned,
				prometheus.GaugeValue,
				float64(*record.OwnedPrivateRepos),
				labels...,
			)
		}

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

	level.Info(c.logger).Log("total", len(orgNames))
}
