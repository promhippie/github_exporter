Enhancement: New metrics and configurations for the workflow collector

1. Added a new metric for the duration, in minutes, of the time since the run was created:
github_workflow_duration_run_created_minutes
2. Added 2 optional configuration options for the workflows exporter:
    1. Query workflows with a specific status (Default to any)
    2. Set the window history (defaults to 12 hours)
3. Added run_id label

https://github.com/promhippie/github_exporter/pull/200
