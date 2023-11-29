Enhancement: Configurable labels for runner metrics

Initially we had a static list of available labels for the runner metrics, with
this change we are getting more labels like the GitHub runner labels as
Prometheus labels while they are configurable to avoid a high cardinality of
Prometheus labels. Now you are able to get labels for `owner`, `id`, `name`,
`os`, `status` and optionally `labels`.

https://github.com/promhippie/github_exporter/issues/277
