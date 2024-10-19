GITHUB_EXPORTER_LOG_LEVEL
: Only log messages with given severity, defaults to `info`

GITHUB_EXPORTER_LOG_PRETTY
: Enable pretty messages for logging, defaults to `false`

GITHUB_EXPORTER_WEB_ADDRESS
: Address to bind the metrics server, defaults to `0.0.0.0:9504`

GITHUB_EXPORTER_WEB_PATH
: Path to bind the metrics server, defaults to `/metrics`

GITHUB_EXPORTER_WEB_PPROF
: Enable pprof debugging for server, defaults to `false`

GITHUB_EXPORTER_WEB_TIMEOUT
: Server metrics endpoint timeout, defaults to `10s`

GITHUB_EXPORTER_WEB_CONFIG
: Path to web-config file

GITHUB_EXPORTER_WEBHOOK_PATH
: Path to webhook target for GitHub, defaults to `/github`

GITHUB_EXPORTER_WEBHOOK_SECRET
: Secret used by GitHub to access webhook

GITHUB_EXPORTER_DATABASE_DSN
: DSN for the database connection

GITHUB_EXPORTER_REQUEST_TIMEOUT
: Timeout requesting GitHub API, defaults to `5s`

GITHUB_EXPORTER_TOKEN
: Access token for the GitHub API, also supports file:// and base64://

GITHUB_EXPORTER_APP_ID
: App ID for the GitHub app, defaults to `0`

GITHUB_EXPORTER_INSTALLATION_ID
: Installation ID for the GitHub app, defaults to `0`

GITHUB_EXPORTER_PRIVATE_KEY
: Private key for the GitHub app, also supports file:// and base64://

GITHUB_EXPORTER_BASE_URL
: URL to access the GitHub Enterprise API

GITHUB_EXPORTER_INSECURE
: Skip TLS verification for GitHub Enterprise, defaults to `false`

GITHUB_EXPORTER_ENTERPRISE, GITHUB_EXPORTER_ENTERPRISES
: Enterprises to scrape metrics from, comma-separated list

GITHUB_EXPORTER_ORG, GITHUB_EXPORTER_ORGS
: Organizations to scrape metrics from, comma-separated list

GITHUB_EXPORTER_REPO, GITHUB_EXPORTER_REPOS
: Repositories to scrape metrics from, comma-separated list

GITHUB_EXPORTER_PER_PAGE
: Number of records per page for API requests, defaults to `500`

GITHUB_EXPORTER_COLLECTOR_ADMIN
: Enable collector for admin stats, defaults to `false`

GITHUB_EXPORTER_COLLECTOR_ORGS
: Enable collector for orgs, defaults to `true`

GITHUB_EXPORTER_COLLECTOR_REPOS
: Enable collector for repos, defaults to `true`

GITHUB_EXPORTER_COLLECTOR_BILLING
: Enable collector for billing, defaults to `false`

GITHUB_EXPORTER_COLLECTOR_WORKFLOWS
: Enable collector for workflows, defaults to `false`

GITHUB_EXPORTER_WORKFLOWS_WINDOW
: History window for querying workflows, defaults to `24h0m0s`

GITHUB_EXPORTER_WORKFLOWS_LABELS
: List of labels used for workflows, comma-separated list, defaults to `owner, repo, workflow, event, name, status, branch, number, run`

GITHUB_EXPORTER_COLLECTOR_WORKFLOW_JOBS
: Enable collector for workflow jobs, defaults to `false`

GITHUB_EXPORTER_WORKFLOW_JOBS_WINDOW
: History window for querying workflow jobs, defaults to `24h0m0s`

GITHUB_EXPORTER_WORKFLOWS_LABELS
: List of labels used for workflow jobs, comma-separated list, defaults to `owner, repo, name, title, branch, sha, run_id, run_attempt, labels`

GITHUB_EXPORTER_COLLECTOR_RUNNERS
: Enable collector for runners, defaults to `false`

GITHUB_EXPORTER_RUNNERS_LABELS
: List of labels used for runners, comma-separated list, defaults to `owner, id, name, os, status`
