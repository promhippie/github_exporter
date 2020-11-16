Feature: Allow insecure TLS via CLI flag
[#19](https://github.com/promhippie/github_exporter/issues/19)

In some cases it can be desirable to ignore certificate errors from the Github API - such as in
the case of connecting to a private instance of Github Enterprise which uses a self-signed cert.

This is exposed via configuration option `GITHUB_EXPORTER_TLS_INSECURE`
