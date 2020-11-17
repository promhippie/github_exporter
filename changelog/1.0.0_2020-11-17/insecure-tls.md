Enhancement: Allow insecure TLS via CLI flag

In some cases it can be desirable to ignore certificate errors from the GitHub
API - such as in the case of connecting to a private instance of GitHub
Enterprise which uses a self-signed cert. This is exposed via the environment
variable `GITHUB_EXPORTER_TLS_INSECURE` and the flag `--github.insecure`.

https://github.com/promhippie/github_exporter/pull/19
