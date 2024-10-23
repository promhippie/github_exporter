Change: Change config and metric names for workflows

We introduced a BREAKING CHANGE by renaming the config variables and metrics
related to the workflows. Previously you had to set `--collector.workflows` or
`GITHUB_EXPORTER_COLLECTOR_WORKFLOWS` to enable the collector, for being
consistent with the new workflow job collector we have renamed them to
`--collector.workflow_runs` and `GITHUB_EXPORTER_COLLECTOR_WORKFLOW_JOBS`, so be
aware about that. Additionally we have also renamed the metrics a tiny bit
matching the same suffix. We renamed them as an example from
`github_workflow_status` to `github_workflow_run_status`.

https://github.com/promhippie/github_exporter/pull/412
