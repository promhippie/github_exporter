Bugfix: Runner ID was being set to run ID for workflow_job

While implementing the workflow job collector we simply attached the wrong
identifier to the runner id which was base on the run ID. Future webhooks will
store the right ID now.

https://github.com/promhippie/github_exporter/pull/417
