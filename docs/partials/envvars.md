GITHUB_EXPORTER_LOG_LEVEL
: Only log messages with given severity, defaults to `info`

GITHUB_EXPORTER_LOG_PRETTY
: Enable pretty messages for logging, defaults to `false`

GITHUB_EXPORTER_WEB_ADDRESS
: Address to bind the metrics server, defaults to `0.0.0.0:9504`

GITHUB_EXPORTER_WEB_PATH
: Path to bind the metrics server, defaults to `/metrics`

GITHUB_EXPORTER_WEB_TIMEOUT
: Server metrics endpoint timeout, defaults to `10s`

GITHUB_EXPORTER_WEB_CONFIG
: Path to web-config file

GITHUB_EXPORTER_REQUEST_TIMEOUT
: Timeout requesting GitHub API, defaults to `5s`

GITHUB_EXPORTER_TOKEN
: Access token for the GitHub API

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

GITHUB_EXPORTER_COLLECTOR_ORGS
: Enable collector for orgs, defaults to `true`

GITHUB_EXPORTER_COLLECTOR_REPOS
: Enable collector for repos, defaults to `true`

GITHUB_EXPORTER_COLLECTOR_ACTIONS
: Enable collector for actions, defaults to `false`

GITHUB_EXPORTER_COLLECTOR_PACKAGES
: Enable collector for packages, defaults to `false`

GITHUB_EXPORTER_COLLECTOR_STORAGE
: Enable collector for storage, defaults to `false`
