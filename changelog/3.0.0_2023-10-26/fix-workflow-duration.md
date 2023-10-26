Bugfix: Use right date for workflow durations

We had used the creation date to detect the workflow duration, this have been
replaced by the run started date to show the right duration time for the
workflow.

https://github.com/promhippie/github_exporter/issues/267
