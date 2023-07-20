Bugfix: Properly handle workflow status results

While a workflow is running the conclusion property provides an empty string, in
order to get a proper value in this case we have changed the metric value to the
status property which offers a useable fallback.

https://github.com/promhippie/github_exporter/issues/230
