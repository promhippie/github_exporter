package exporter

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v60/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/store"
)

// AdminCollector collects metrics about the servers.
type AdminCollector struct {
	client   *github.Client
	logger   log.Logger
	db       store.Store
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	ReposTotal       *prometheus.Desc
	ReposRoot        *prometheus.Desc
	ReposFork        *prometheus.Desc
	ReposOrg         *prometheus.Desc
	ReposTotalPushes *prometheus.Desc
	ReposTotalWikis  *prometheus.Desc

	HooksTotal    *prometheus.Desc
	HooksActive   *prometheus.Desc
	HooksInactive *prometheus.Desc

	PagesTotal *prometheus.Desc

	OrgsTotal        *prometheus.Desc
	OrgsDisabled     *prometheus.Desc
	OrgsTotalTeams   *prometheus.Desc
	OrgsTotalMembers *prometheus.Desc

	UsersTotal     *prometheus.Desc
	UsersAdmin     *prometheus.Desc
	UsersSuspended *prometheus.Desc

	PullsTotal       *prometheus.Desc
	PullsMerged      *prometheus.Desc
	PullsMergeable   *prometheus.Desc
	PullsUnmergeable *prometheus.Desc

	IssuesTotal  *prometheus.Desc
	IssuesOpen   *prometheus.Desc
	IssuesClosed *prometheus.Desc

	MilestonesTotal  *prometheus.Desc
	MilestonesOpen   *prometheus.Desc
	MilestonesClosed *prometheus.Desc

	GistsTotal   *prometheus.Desc
	GistsPrivate *prometheus.Desc
	GistsPublic  *prometheus.Desc

	CommentsCommit      *prometheus.Desc
	CommentsGist        *prometheus.Desc
	CommentsIssue       *prometheus.Desc
	CommentsPullRequest *prometheus.Desc
}

// NewAdminCollector returns a new AdminCollector.
func NewAdminCollector(logger log.Logger, client *github.Client, db store.Store, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *AdminCollector {
	if failures != nil {
		failures.WithLabelValues("admin").Add(0)
	}

	labels := []string{}
	return &AdminCollector{
		client:   client,
		logger:   log.With(logger, "collector", "admin"),
		db:       db,
		failures: failures,
		duration: duration,
		config:   cfg,

		ReposTotal: prometheus.NewDesc(
			"github_admin_repos_total",
			"Total number of repositories",
			labels,
			nil,
		),
		ReposRoot: prometheus.NewDesc(
			"github_admin_repos_root",
			"Number of root repositories",
			labels,
			nil,
		),
		ReposFork: prometheus.NewDesc(
			"github_admin_repos_fork",
			"Number of fork repositories",
			labels,
			nil,
		),
		ReposOrg: prometheus.NewDesc(
			"github_admin_repos_org",
			"Number of organization repos",
			labels,
			nil,
		),
		ReposTotalPushes: prometheus.NewDesc(
			"github_admin_repos_pushes_total",
			"Total number of pushes",
			labels,
			nil,
		),
		ReposTotalWikis: prometheus.NewDesc(
			"github_admin_repos_wikis_total",
			"Total number of wikis",
			labels,
			nil,
		),

		HooksTotal: prometheus.NewDesc(
			"github_admin_hooks_total",
			"Total number of hooks",
			labels,
			nil,
		),
		HooksActive: prometheus.NewDesc(
			"github_admin_hooks_active",
			"Number of active hooks",
			labels,
			nil,
		),
		HooksInactive: prometheus.NewDesc(
			"github_admin_hooks_inactive",
			"Number of inactive hooks",
			labels,
			nil,
		),

		PagesTotal: prometheus.NewDesc(
			"github_admin_pages_total",
			"Total number of pages",
			labels,
			nil,
		),

		OrgsTotal: prometheus.NewDesc(
			"github_admin_orgs_total",
			"Total number of organizations",
			labels,
			nil,
		),
		OrgsDisabled: prometheus.NewDesc(
			"github_admin_orgs_disabled",
			"Number of disabled organizations",
			labels,
			nil,
		),
		OrgsTotalTeams: prometheus.NewDesc(
			"github_admin_orgs_teams",
			"Number of organization teams",
			labels,
			nil,
		),
		OrgsTotalMembers: prometheus.NewDesc(
			"github_admin_orgs_members",
			"Number of organization team members",
			labels,
			nil,
		),

		UsersTotal: prometheus.NewDesc(
			"github_admin_users_total",
			"Total number of users",
			labels,
			nil,
		),
		UsersAdmin: prometheus.NewDesc(
			"github_admin_users_admin",
			"Number of admin users",
			labels,
			nil,
		),
		UsersSuspended: prometheus.NewDesc(
			"github_admin_users_suspended",
			"Number of suspended users",
			labels,
			nil,
		),

		PullsTotal: prometheus.NewDesc(
			"github_admin_pulls_total",
			"Total number of pull requests",
			labels,
			nil,
		),
		PullsMerged: prometheus.NewDesc(
			"github_admin_pulls_merged",
			"Number of merged pull requests",
			labels,
			nil,
		),
		PullsMergeable: prometheus.NewDesc(
			"github_admin_pulls_mergeable",
			"Number of mergeable pull requests",
			labels,
			nil,
		),
		PullsUnmergeable: prometheus.NewDesc(
			"github_admin_pulls_unmergeable",
			"Number of unmergeable pull requests",
			labels,
			nil,
		),

		IssuesTotal: prometheus.NewDesc(
			"github_admin_issues_total",
			"Total number of issues",
			labels,
			nil,
		),
		IssuesOpen: prometheus.NewDesc(
			"github_admin_issues_open",
			"Number of open issues",
			labels,
			nil,
		),
		IssuesClosed: prometheus.NewDesc(
			"github_admin_issues_closed",
			"Number of closed issues",
			labels,
			nil,
		),

		MilestonesTotal: prometheus.NewDesc(
			"github_admin_milestones_total",
			"Total number of milestones",
			labels,
			nil,
		),
		MilestonesOpen: prometheus.NewDesc(
			"github_admin_milestones_open",
			"Number of open milestones",
			labels,
			nil,
		),
		MilestonesClosed: prometheus.NewDesc(
			"github_admin_milestones_closed",
			"Number of closed milestones",
			labels,
			nil,
		),

		GistsTotal: prometheus.NewDesc(
			"github_admin_gists_total",
			"Total number of gists",
			labels,
			nil,
		),
		GistsPrivate: prometheus.NewDesc(
			"github_admin_gists_private",
			"Number of private gists",
			labels,
			nil,
		),
		GistsPublic: prometheus.NewDesc(
			"github_admin_gists_public",
			"Number of public gists",
			labels,
			nil,
		),

		CommentsCommit: prometheus.NewDesc(
			"github_admin_comments_commit",
			"Number of commit comments",
			labels,
			nil,
		),
		CommentsGist: prometheus.NewDesc(
			"github_admin_comments_gist",
			"Number of gist comments",
			labels,
			nil,
		),
		CommentsIssue: prometheus.NewDesc(
			"github_admin_comments_issue",
			"Number of issue comments",
			labels,
			nil,
		),
		CommentsPullRequest: prometheus.NewDesc(
			"github_admin_comments_pull_request",
			"Number of pull request comments",
			labels,
			nil,
		),
	}
}

// Metrics simply returns the list metric descriptors for generating a documentation.
func (c *AdminCollector) Metrics() []*prometheus.Desc {
	return []*prometheus.Desc{
		c.ReposTotal,
		c.ReposRoot,
		c.ReposFork,
		c.ReposOrg,
		c.ReposTotalPushes,
		c.ReposTotalWikis,
		c.HooksTotal,
		c.HooksActive,
		c.HooksInactive,
		c.PagesTotal,
		c.OrgsTotal,
		c.OrgsDisabled,
		c.OrgsTotalTeams,
		c.OrgsTotalMembers,
		c.UsersTotal,
		c.UsersAdmin,
		c.UsersSuspended,
		c.PullsTotal,
		c.PullsMerged,
		c.PullsMergeable,
		c.PullsUnmergeable,
		c.IssuesTotal,
		c.IssuesOpen,
		c.IssuesClosed,
		c.MilestonesTotal,
		c.MilestonesOpen,
		c.MilestonesClosed,
		c.GistsTotal,
		c.GistsPrivate,
		c.GistsPublic,
		c.CommentsCommit,
		c.CommentsGist,
		c.CommentsIssue,
		c.CommentsPullRequest,
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *AdminCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ReposTotal
	ch <- c.ReposRoot
	ch <- c.ReposFork
	ch <- c.ReposOrg
	ch <- c.ReposTotalPushes
	ch <- c.ReposTotalWikis
	ch <- c.HooksTotal
	ch <- c.HooksActive
	ch <- c.HooksInactive
	ch <- c.PagesTotal
	ch <- c.OrgsTotal
	ch <- c.OrgsDisabled
	ch <- c.OrgsTotalTeams
	ch <- c.OrgsTotalMembers
	ch <- c.UsersTotal
	ch <- c.UsersAdmin
	ch <- c.UsersSuspended
	ch <- c.PullsTotal
	ch <- c.PullsMerged
	ch <- c.PullsMergeable
	ch <- c.PullsUnmergeable
	ch <- c.IssuesTotal
	ch <- c.IssuesOpen
	ch <- c.IssuesClosed
	ch <- c.MilestonesTotal
	ch <- c.MilestonesOpen
	ch <- c.MilestonesClosed
	ch <- c.GistsTotal
	ch <- c.GistsPrivate
	ch <- c.GistsPublic
	ch <- c.CommentsCommit
	ch <- c.CommentsGist
	ch <- c.CommentsIssue
	ch <- c.CommentsPullRequest
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AdminCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	now := time.Now()
	record, resp, err := c.client.Admin.GetAdminStats(ctx)
	c.duration.WithLabelValues("admin").Observe(time.Since(now).Seconds())
	defer closeBody(resp)

	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to fetch admin stats",
			"err", err,
		)

		c.failures.WithLabelValues("admin").Inc()
		return
	}

	level.Debug(c.logger).Log(
		"msg", "Fetched admin stats",
	)

	labels := []string{}

	ch <- prometheus.MustNewConstMetric(
		c.ReposTotal,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetTotalRepos()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReposRoot,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetRootRepos()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReposFork,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetForkRepos()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReposOrg,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetOrgRepos()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReposTotalPushes,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetTotalPushes()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ReposTotalWikis,
		prometheus.GaugeValue,
		float64(record.GetRepos().GetTotalWikis()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.HooksTotal,
		prometheus.GaugeValue,
		float64(record.GetHooks().GetTotalHooks()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.HooksActive,
		prometheus.GaugeValue,
		float64(record.GetHooks().GetActiveHooks()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.HooksInactive,
		prometheus.GaugeValue,
		float64(record.GetHooks().GetInactiveHooks()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagesTotal,
		prometheus.GaugeValue,
		float64(record.GetPages().GetTotalPages()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OrgsTotal,
		prometheus.GaugeValue,
		float64(record.GetOrgs().GetTotalOrgs()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OrgsDisabled,
		prometheus.GaugeValue,
		float64(record.GetOrgs().GetDisabledOrgs()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OrgsTotalTeams,
		prometheus.GaugeValue,
		float64(record.GetOrgs().GetTotalTeams()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.OrgsTotalMembers,
		prometheus.GaugeValue,
		float64(record.GetOrgs().GetTotalTeamMembers()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.UsersTotal,
		prometheus.GaugeValue,
		float64(record.GetUsers().GetTotalUsers()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.UsersAdmin,
		prometheus.GaugeValue,
		float64(record.GetUsers().GetAdminUsers()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.UsersSuspended,
		prometheus.GaugeValue,
		float64(record.GetUsers().GetSuspendedUsers()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PullsTotal,
		prometheus.GaugeValue,
		float64(record.GetPulls().GetTotalPulls()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PullsMerged,
		prometheus.GaugeValue,
		float64(record.GetPulls().GetMergedPulls()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PullsMergeable,
		prometheus.GaugeValue,
		float64(record.GetPulls().GetMergablePulls()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PullsUnmergeable,
		prometheus.GaugeValue,
		float64(record.GetPulls().GetUnmergablePulls()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.IssuesTotal,
		prometheus.GaugeValue,
		float64(record.GetIssues().GetTotalIssues()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.IssuesOpen,
		prometheus.GaugeValue,
		float64(record.GetIssues().GetOpenIssues()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.IssuesClosed,
		prometheus.GaugeValue,
		float64(record.GetIssues().GetClosedIssues()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MilestonesTotal,
		prometheus.GaugeValue,
		float64(record.GetMilestones().GetTotalMilestones()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MilestonesOpen,
		prometheus.GaugeValue,
		float64(record.GetMilestones().GetOpenMilestones()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.MilestonesClosed,
		prometheus.GaugeValue,
		float64(record.GetMilestones().GetClosedMilestones()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GistsTotal,
		prometheus.GaugeValue,
		float64(record.GetGists().GetTotalGists()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GistsPrivate,
		prometheus.GaugeValue,
		float64(record.GetGists().GetPrivateGists()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.GistsPublic,
		prometheus.GaugeValue,
		float64(record.GetGists().GetPublicGists()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommentsCommit,
		prometheus.GaugeValue,
		float64(record.GetComments().GetTotalCommitComments()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommentsGist,
		prometheus.GaugeValue,
		float64(record.GetComments().GetTotalGistComments()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommentsIssue,
		prometheus.GaugeValue,
		float64(record.GetComments().GetTotalIssueComments()),
		labels...,
	)

	ch <- prometheus.MustNewConstMetric(
		c.CommentsPullRequest,
		prometheus.GaugeValue,
		float64(record.GetComments().GetTotalPullRequestComments()),
		labels...,
	)
}
