Enhancement: Add metrics for GitHub workflows

We've added new metrics for the observability of GitHub workflows/actions, they
are disabled by default because it could result in a high cardinality of the
labels.

https://github.com/promhippie/github_exporter/issues/123
