Change: Renamed valid_response metric to up metric

The previous metric `github_valid_response` doesn't match the Prometheus
conventions, so it have been renamed to `github_up` which properly signals if
the exporter can gather metrics or not.

https://github.com/promhippie/github_exporter/issues/2
