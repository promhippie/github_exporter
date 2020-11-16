# Changelog for unreleased

The following sections list the changes for unreleased.

## Summary

 * Allow insecure TLS via CLI flag
   [#19](https://github.com/promhippie/github_exporter/issues/19)

## Details

 * In some cases it can be desirable to ignore certificate errors from the Github API - such as in
   the case of connecting to a private instance of Github Enterprise which uses a self-signed cert.

   This is exposed via configuration option `GITHUB_EXPORTER_TLS_INSECURE`


## Summary

 * Chg #12: Refactor structure and integrate more metrics

## Details

 * Change #12: Refactor structure and integrate more metrics

   The structure of the repository should get overhauled and beside that we should gather more
   metrics per organization and per repository. Additionally we should also add support for
   GitHub Enterprise as requested within
   [#10](https://github.com/promhippie/github_exporter/issues/10).

   https://github.com/promhippie/github_exporter/issues/12


# Changelog for 0.2.0

The following sections list the changes for 0.2.0.

## Summary

 * Chg #4: Enforce a repo or an org flag
 * Chg #2: Renamed valid_response metric to up metric

## Details

 * Change #4: Enforce a repo or an org flag

   The exporter requires at least one organization or repository to work properly, integrated a
   check that something have been set when launching the exporter.

   https://github.com/promhippie/github_exporter/issues/4

 * Change #2: Renamed valid_response metric to up metric

   The previous metric `github_valid_response` doesn't match the Prometheus conventions, so
   it have been renamed to `github_up` which properly signals if the exporter can gather metrics
   or not.

   https://github.com/promhippie/github_exporter/issues/2


# Changelog for 0.1.0

The following sections list the changes for 0.1.0.

## Summary

 * Chg #11: Initial release of basic version

## Details

 * Change #11: Initial release of basic version

   Just prepared an initial basic version which could be released to the public.

   https://github.com/promhippie/github_exporter/issues/11


