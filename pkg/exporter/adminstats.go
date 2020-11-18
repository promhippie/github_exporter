package exporter

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/go-github/v32/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/promhippie/github_exporter/pkg/config"
)

// AdminStatsCollector collects metrics about GHE instances
type AdminStatsCollector struct {
	client   *github.Client
	logger   log.Logger
	failures *prometheus.CounterVec
	duration *prometheus.HistogramVec
	config   config.Target

	// repos
	ReposTotal  *prometheus.Desc
	ReposRoot   *prometheus.Desc
	ReposForked *prometheus.Desc
	ReposOrg    *prometheus.Desc
	PushedTotal *prometheus.Desc
	WikisTotal  *prometheus.Desc
	// hooks
	HooksTotal    *prometheus.Desc
	HooksActive   *prometheus.Desc
	HooksInactive *prometheus.Desc
	// pages
	PagesTotal *prometheus.Desc
	// orgs
	OrgsTotal            *prometheus.Desc
	OrgsDisabled         *prometheus.Desc
	OrgsTeamsTotal       *prometheus.Desc
	OrgsTeamMembersTotal *prometheus.Desc
	// users
	UsersTotal     *prometheus.Desc
	UsersAdmin     *prometheus.Desc
	UsersSuspended *prometheus.Desc
	// pulls
	PullsTotal       *prometheus.Desc
	PullsMerged      *prometheus.Desc
	PullsMergeable   *prometheus.Desc
	PullsUnmergeable *prometheus.Desc
	// issues
	IssuesTotal  *prometheus.Desc
	IssuesOpen   *prometheus.Desc
	IssuesClosed *prometheus.Desc
	// milestones
	MilestonesTotal  *prometheus.Desc
	MilestonesOpen   *prometheus.Desc
	MilestonesClosed *prometheus.Desc
	// gists
	GistsTotal   *prometheus.Desc
	GistsPrivate *prometheus.Desc
	GistsPublic  *prometheus.Desc
	// comments
	CommentsCommitTotal      *prometheus.Desc
	CommentsGistTotal        *prometheus.Desc
	CommentsIssueTotal       *prometheus.Desc
	CommentsPullRequestTotal *prometheus.Desc
}

// NewAdminStatsCollector returns a new AdminStatsCollector.
func NewAdminStatsCollector(logger log.Logger, client *github.Client, failures *prometheus.CounterVec, duration *prometheus.HistogramVec, cfg config.Target) *AdminStatsCollector {
	failures.WithLabelValues("org").Add(0)

	return &AdminStatsCollector{
		client:   client,
		logger:   log.With(logger, "collector", "adminstats"),
		failures: failures,
		duration: duration,
		config:   cfg,

		ReposTotal:               prometheus.NewDesc("github_repos", "Number of repos", nil, nil),
		ReposRoot:                prometheus.NewDesc("github_root_repos", "Number of root repos", nil, nil),
		ReposForked:              prometheus.NewDesc("github_forked_repos", "Number of forked repos", nil, nil),
		ReposOrg:                 prometheus.NewDesc("github_org_repos", "Number of org repos", nil, nil),
		PushedTotal:              prometheus.NewDesc("github_pushes", "Number of pushes", nil, nil),
		WikisTotal:               prometheus.NewDesc("github_wikis", "Number of wikis", nil, nil),
		HooksTotal:               prometheus.NewDesc("github_hooks", "Number of hooks", nil, nil),
		HooksActive:              prometheus.NewDesc("github_active_hooks", "Number of active hooks", nil, nil),
		HooksInactive:            prometheus.NewDesc("github_inactive_hooks", "Number of inactive hooks", nil, nil),
		PagesTotal:               prometheus.NewDesc("github_pages", "Number of pages", nil, nil),
		OrgsTotal:                prometheus.NewDesc("github_orgs", "Number of organizations", nil, nil),
		OrgsDisabled:             prometheus.NewDesc("github_disabled_orgs", "Number of disabled organizations", nil, nil),
		OrgsTeamsTotal:           prometheus.NewDesc("github_org_teams", "Number of organization teams", nil, nil),
		OrgsTeamMembersTotal:     prometheus.NewDesc("github_org_team_members", "Number of organization team members", nil, nil),
		UsersTotal:               prometheus.NewDesc("github_users", "Number of users", nil, nil),
		UsersAdmin:               prometheus.NewDesc("github_admin_users", "Number of admin users", nil, nil),
		UsersSuspended:           prometheus.NewDesc("github_suspended_users", "Number of suspended users", nil, nil),
		PullsTotal:               prometheus.NewDesc("github_pulls", "Number of pulls", nil, nil),
		PullsMerged:              prometheus.NewDesc("github_merged_pulls", "Number of merged pulls", nil, nil),
		PullsMergeable:           prometheus.NewDesc("github_mergeable_pulls", "Number of mergeable pulls", nil, nil),
		PullsUnmergeable:         prometheus.NewDesc("github_unmergeable_pulls", "Number of unmergeable pulls", nil, nil),
		IssuesTotal:              prometheus.NewDesc("github_issues", "Number of issues", nil, nil),
		IssuesOpen:               prometheus.NewDesc("github_open_issues", "Number of open issues", nil, nil),
		IssuesClosed:             prometheus.NewDesc("github_closed_issues", "Number of closed issues", nil, nil),
		MilestonesTotal:          prometheus.NewDesc("github_milestones", "Number of milestones", nil, nil),
		MilestonesOpen:           prometheus.NewDesc("github_open_milestones", "Number of open milestones", nil, nil),
		MilestonesClosed:         prometheus.NewDesc("github_closed_milestones", "Number of closed milestones", nil, nil),
		GistsTotal:               prometheus.NewDesc("github_gists", "Number of gists", nil, nil),
		GistsPrivate:             prometheus.NewDesc("github_private_gists", "Number of private gists", nil, nil),
		GistsPublic:              prometheus.NewDesc("github_public_gists", "Number of public gists", nil, nil),
		CommentsCommitTotal:      prometheus.NewDesc("github_commit_comments", "Number of commit comments", nil, nil),
		CommentsGistTotal:        prometheus.NewDesc("github_gist_comments", "Number of gist comments", nil, nil),
		CommentsIssueTotal:       prometheus.NewDesc("github_issue_comments", "Number of issue comments", nil, nil),
		CommentsPullRequestTotal: prometheus.NewDesc("github_pullrequest_comments", "Number of pull request comments", nil, nil),
	}
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector.
func (c *AdminStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ReposTotal
	ch <- c.ReposRoot
	ch <- c.ReposForked
	ch <- c.ReposOrg
	ch <- c.PushedTotal
	ch <- c.WikisTotal
	ch <- c.HooksTotal
	ch <- c.HooksActive
	ch <- c.HooksInactive
	ch <- c.PagesTotal
	ch <- c.OrgsTotal
	ch <- c.OrgsDisabled
	ch <- c.OrgsTeamsTotal
	ch <- c.OrgsTeamMembersTotal
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
	ch <- c.CommentsCommitTotal
	ch <- c.CommentsGistTotal
	ch <- c.CommentsIssueTotal
	ch <- c.CommentsPullRequestTotal
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AdminStatsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	now := time.Now()
	record, _, err := c.client.Admin.GetAdminStats(ctx)
	c.duration.WithLabelValues("adminstats").Observe(time.Since(now).Seconds())

	if err != nil {
		level.Error(c.logger).Log(
			"msg", "Failed to fetch Private GHE stats",
			"err", err,
		)

		c.failures.WithLabelValues("adminstats").Inc()
		return
	}

	// repos
	ch <- prometheus.MustNewConstMetric(c.ReposTotal, prometheus.GaugeValue, float64(*record.Repos.TotalRepos))
	ch <- prometheus.MustNewConstMetric(c.ReposRoot, prometheus.GaugeValue, float64(*record.Repos.RootRepos))
	ch <- prometheus.MustNewConstMetric(c.ReposForked, prometheus.GaugeValue, float64(*record.Repos.ForkRepos))
	ch <- prometheus.MustNewConstMetric(c.ReposOrg, prometheus.GaugeValue, float64(*record.Repos.OrgRepos))
	ch <- prometheus.MustNewConstMetric(c.PushedTotal, prometheus.GaugeValue, float64(*record.Repos.TotalPushes))
	ch <- prometheus.MustNewConstMetric(c.WikisTotal, prometheus.GaugeValue, float64(*record.Repos.TotalWikis))

	// hooks
	ch <- prometheus.MustNewConstMetric(c.HooksTotal, prometheus.GaugeValue, float64(*record.Hooks.TotalHooks))
	ch <- prometheus.MustNewConstMetric(c.HooksActive, prometheus.GaugeValue, float64(*record.Hooks.ActiveHooks))
	ch <- prometheus.MustNewConstMetric(c.HooksInactive, prometheus.GaugeValue, float64(*record.Hooks.InactiveHooks))

	// pages
	ch <- prometheus.MustNewConstMetric(c.PagesTotal, prometheus.GaugeValue, float64(*record.Pages.TotalPages))

	// orgs
	ch <- prometheus.MustNewConstMetric(c.OrgsTotal, prometheus.GaugeValue, float64(*record.Orgs.TotalOrgs))
	ch <- prometheus.MustNewConstMetric(c.OrgsDisabled, prometheus.GaugeValue, float64(*record.Orgs.DisabledOrgs))
	ch <- prometheus.MustNewConstMetric(c.OrgsTeamsTotal, prometheus.GaugeValue, float64(*record.Orgs.TotalTeams))
	ch <- prometheus.MustNewConstMetric(c.OrgsTeamMembersTotal, prometheus.GaugeValue, float64(*record.Orgs.TotalTeamMembers))

	// users
	ch <- prometheus.MustNewConstMetric(c.UsersTotal, prometheus.GaugeValue, float64(*record.Users.TotalUsers))
	ch <- prometheus.MustNewConstMetric(c.UsersAdmin, prometheus.GaugeValue, float64(*record.Users.AdminUsers))
	ch <- prometheus.MustNewConstMetric(c.UsersSuspended, prometheus.GaugeValue, float64(*record.Users.SuspendedUsers))

	// pulls
	ch <- prometheus.MustNewConstMetric(c.PullsTotal, prometheus.GaugeValue, float64(*record.Pulls.TotalPulls))
	ch <- prometheus.MustNewConstMetric(c.PullsMerged, prometheus.GaugeValue, float64(*record.Pulls.MergedPulls))
	ch <- prometheus.MustNewConstMetric(c.PullsMergeable, prometheus.GaugeValue, float64(*record.Pulls.MergablePulls))
	ch <- prometheus.MustNewConstMetric(c.PullsUnmergeable, prometheus.GaugeValue, float64(*record.Pulls.UnmergablePulls))

	// issues
	ch <- prometheus.MustNewConstMetric(c.IssuesTotal, prometheus.GaugeValue, float64(*record.Issues.TotalIssues))
	ch <- prometheus.MustNewConstMetric(c.IssuesOpen, prometheus.GaugeValue, float64(*record.Issues.OpenIssues))
	ch <- prometheus.MustNewConstMetric(c.IssuesClosed, prometheus.GaugeValue, float64(*record.Issues.ClosedIssues))

	// milestones
	ch <- prometheus.MustNewConstMetric(c.MilestonesTotal, prometheus.GaugeValue, float64(*record.Milestones.TotalMilestones))
	ch <- prometheus.MustNewConstMetric(c.MilestonesOpen, prometheus.GaugeValue, float64(*record.Milestones.OpenMilestones))
	ch <- prometheus.MustNewConstMetric(c.MilestonesClosed, prometheus.GaugeValue, float64(*record.Milestones.ClosedMilestones))

	// gists
	ch <- prometheus.MustNewConstMetric(c.GistsTotal, prometheus.GaugeValue, float64(*record.Gists.TotalGists))
	ch <- prometheus.MustNewConstMetric(c.GistsPrivate, prometheus.GaugeValue, float64(*record.Gists.PrivateGists))
	ch <- prometheus.MustNewConstMetric(c.GistsPublic, prometheus.GaugeValue, float64(*record.Gists.PublicGists))

	// comments
	ch <- prometheus.MustNewConstMetric(c.CommentsCommitTotal, prometheus.GaugeValue, float64(*record.Comments.TotalCommitComments))
	ch <- prometheus.MustNewConstMetric(c.CommentsGistTotal, prometheus.GaugeValue, float64(*record.Comments.TotalGistComments))
	ch <- prometheus.MustNewConstMetric(c.CommentsIssueTotal, prometheus.GaugeValue, float64(*record.Comments.TotalIssueComments))
	ch <- prometheus.MustNewConstMetric(c.CommentsPullRequestTotal, prometheus.GaugeValue, float64(*record.Comments.TotalPullRequestComments))
}
