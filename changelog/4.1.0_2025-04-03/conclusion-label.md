Enhancement: Optional conclusion label for workflow jobs

We added another optional label which includes the `conclusion` value as an
extra label as this could be different than the `status` value. This label got
to be enabled by the `--collector.workflow_jobs.labels` flag or the
`GITHUB_EXPORTER_WORKFLOW_JOBS_LABELS` environment variable.

https://github.com/promhippie/github_exporter/pull/469
