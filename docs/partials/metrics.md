github_admin_comments_commit{}
: Number of commit comments

github_admin_comments_gist{}
: Number of gist comments

github_admin_comments_issue{}
: Number of issue comments

github_admin_comments_pull_request{}
: Number of pull request comments

github_admin_gists_private{}
: Number of private gists

github_admin_gists_public{}
: Number of public gists

github_admin_gists_total{}
: Total number of gists

github_admin_hooks_active{}
: Number of active hooks

github_admin_hooks_inactive{}
: Number of inactive hooks

github_admin_hooks_total{}
: Total number of hooks

github_admin_issues_closed{}
: Number of closed issues

github_admin_issues_open{}
: Number of open issues

github_admin_issues_total{}
: Total number of issues

github_admin_milestones_closed{}
: Number of closed milestones

github_admin_milestones_open{}
: Number of open milestones

github_admin_milestones_total{}
: Total number of milestones

github_admin_orgs_disabled{}
: Number of disabled organizations

github_admin_orgs_members{}
: Number of organization team members

github_admin_orgs_teams{}
: Number of organization teams

github_admin_orgs_total{}
: Total number of organizations

github_admin_pages_total{}
: Total number of pages

github_admin_pulls_mergeable{}
: Number of mergeable pull requests

github_admin_pulls_merged{}
: Number of merged pull requests

github_admin_pulls_total{}
: Total number of pull requests

github_admin_pulls_unmergeable{}
: Number of unmergeable pull requests

github_admin_repos_fork{}
: Number of fork repositories

github_admin_repos_org{}
: Number of organization repos

github_admin_repos_pushes_total{}
: Total number of pushes

github_admin_repos_root{}
: Number of root repositories

github_admin_repos_total{}
: Total number of repositories

github_admin_repos_wikis_total{}
: Total number of wikis

github_admin_users_admin{}
: Number of admin users

github_admin_users_suspended{}
: Number of suspended users

github_admin_users_total{}
: Total number of users

github_billing_usage_cost_discount{type, name, product, sku, unit_type, date, organization, repository}
: Discount amount applied to GitHub product usage (v5.0.0+)

github_billing_usage_cost_gross{type, name, product, sku, unit_type, date, organization, repository}
: Gross cost before discounts for GitHub product usage (v5.0.0+)

github_billing_usage_cost_net{type, name, product, sku, unit_type, date, organization, repository}
: Net cost after discounts for GitHub product usage - actual charges (v5.0.0+)

github_billing_usage_price_per_unit{type, name, product, sku, unit_type, date, organization, repository}
: Price per unit for GitHub product usage (v5.0.0+)

github_billing_usage_quantity{type, name, product, sku, unit_type, date, organization, repository}
: Usage quantity for GitHub products with repository-level attribution (v5.0.0+)

github_org_collaborators{name}
: Number of collaborators within org

github_org_create_timestamp{name}
: Timestamp of the creation of org

github_org_disk_usage{name}
: Used diskspace by the org

github_org_filled_seats{name}
: Filled seats for org

github_org_followers{name}
: Number of followers for org

github_org_following{name}
: Number of following other users by org

github_org_private_gists{name}
: Number of private gists from org

github_org_private_repos_owned{name}
: Owned private repositories by org

github_org_private_repos_total{name}
: Total amount of private repositories

github_org_public_gists{name}
: Number of public gists from org

github_org_public_repos{name}
: Number of public repositories from org

github_org_seats{name}
: Seats for org

github_org_updated_timestamp{name}
: Timestamp of the last modification of org

github_repo_allow_merge_commit{owner, name}
: Show if this repository allows merge commits

github_repo_allow_rebase_merge{owner, name}
: Show if this repository allows rebase merges

github_repo_allow_squash_merge{owner, name}
: Show if this repository allows squash merges

github_repo_archived{owner, name}
: Show if this repository have been archived

github_repo_created_timestamp{owner, name}
: Timestamp of the creation of repo

github_repo_forked{owner, name}
: Show if this repository is a forked repository

github_repo_forks{owner, name}
: How often has this repository been forked

github_repo_has_downloads{owner, name}
: Show if this repository got downloads enabled

github_repo_has_issues{owner, name}
: Show if this repository got issues enabled

github_repo_has_pages{owner, name}
: Show if this repository got pages enabled

github_repo_has_projects{owner, name}
: Show if this repository got projects enabled

github_repo_has_wiki{owner, name}
: Show if this repository got wiki enabled

github_repo_issues{owner, name}
: Number of open issues on this repository

github_repo_network{owner, name}
: Number of repositories in the network

github_repo_private{owner, name}
: Show iof this repository is private

github_repo_pushed_timestamp{owner, name}
: Timestamp of the last push to repo

github_repo_size{owner, name}
: Size of the repository content

github_repo_stargazers{owner, name}
: Number of stargazers on this repository

github_repo_subscribers{owner, name}
: Number of subscribers on this repository

github_repo_updated_timestamp{owner, name}
: Timestamp of the last modification of repo

github_repo_watchers{owner, name}
: Number of watchers on this repository

github_request_duration_seconds{collector}
: Histogram of latencies for requests to the api per collector

github_request_failures_total{collector}
: Total number of failed requests to the api per collector

github_runner_enterprise_busy{owner, id, name, os, status}
: 1 if the runner is busy, 0 otherwise

github_runner_enterprise_online{owner, id, name, os, status}
: Static metrics of runner is online or not

github_runner_org_busy{owner, id, name, os, status}
: 1 if the runner is busy, 0 otherwise

github_runner_org_online{owner, id, name, os, status}
: Static metrics of runner is online or not

github_runner_repo_busy{owner, id, name, os, status}
: 1 if the runner is busy, 0 otherwise

github_runner_repo_online{owner, id, name, os, status}
: Static metrics of runner is online or not

github_workflow_job_created_timestamp{owner, repo, name, title, branch, sha, identifier, run_id, run_attempt, labels, runner_id, runner_name, runner_group_id, runner_group_name, workflow_name, conclusion}
: Timestamp when the workflow job have been created

github_workflow_job_duration_ms{owner, repo, name, title, branch, sha, identifier, run_id, run_attempt, labels, runner_id, runner_name, runner_group_id, runner_group_name, workflow_name, conclusion}
: Duration of workflow runs

github_workflow_job_duration_run_created_minutes{owner, repo, name, title, branch, sha, identifier, run_id, run_attempt, labels, runner_id, runner_name, runner_group_id, runner_group_name, workflow_name, conclusion}
: Duration since the workflow run creation time in minutes

github_workflow_job_started_timestamp{owner, repo, name, title, branch, sha, identifier, run_id, run_attempt, labels, runner_id, runner_name, runner_group_id, runner_group_name, workflow_name, conclusion}
: Timestamp when the workflow job have been started

github_workflow_job_status{owner, repo, name, title, branch, sha, identifier, run_id, run_attempt, labels, runner_id, runner_name, runner_group_id, runner_group_name, workflow_name, conclusion}
: Status of workflow jobs

github_workflow_run_created_timestamp{owner, repo, workflow, event, name, status, branch, number, run}
: Timestamp when the workflow run have been created

github_workflow_run_duration_ms{owner, repo, workflow, event, name, status, branch, number, run}
: Duration of workflow runs

github_workflow_run_duration_run_created_minutes{owner, repo, workflow, event, name, status, branch, number, run}
: Duration since the workflow run creation time in minutes

github_workflow_run_started_timestamp{owner, repo, workflow, event, name, status, branch, number, run}
: Timestamp when the workflow run have been started

github_workflow_run_status{owner, repo, workflow, event, name, status, branch, number, run}
: Status of workflow runs

github_workflow_run_updated_timestamp{owner, repo, workflow, event, name, status, branch, number, run}
: Timestamp when the workflow run have been updated
