Enhancement: New metrics and configs for workflow collector

We have added a new metric for the duration of the time since the workflow run
was created, defined in minutes. Beside that we have added two additional
configurations to query the workflows for a specific status and you are able to
define a different timeframe than 12 hours now. Finally we have also added the
run ID to the labels.

https://github.com/promhippie/github_exporter/pull/200
https://github.com/promhippie/github_exporter/pull/214
