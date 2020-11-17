Enhancement: Add flag to set /metrics endpoint request timeout

When pulling a lot of data from the GitHub API, in some cases the default 10s 
timeout on the `/metrics` endpoint can be insufficient. This option allows the
timeout to be configured via `GITHUB_EXPORTER_SERVER_TIMEOUT`

https://github.com/promhippie/github_exporter/pull/20
