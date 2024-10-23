Change: New metrics and configs for workflow job collector

We have added a new metric for the duration of the time since the workflow job
was created, defined in minutes. Beside that we have added two additional
configurations to query the workflows for a specific status and you are able to
define a different timeframe than 12 hours now.

https://github.com/promhippie/github_exporter/pull/405
