# Changelog for 1.1.1

The following sections list the changes for 1.1.1.

## Summary

 * Fix #75: Fixed wildcard matching for private repos

## Details

 * Bugfix #75: Fixed wildcard matching for private repos

   We have fixed the ability to match by globbing/wildcard for private repositories. So far the
   private repositories have been missing with the latest globbing/wildcard matching.

   https://github.com/promhippie/github_exporter/issues/75


# Changelog for 1.1.0

The following sections list the changes for 1.1.0.

## Summary

 * Fix #67: Fixed typecasts within billing API
 * Chg #45: Change docker image name
 * Chg #42: Drop darwin/386 release builds
 * Chg #46: Generate metrics documentation
 * Chg #71: Integrate standard web config
 * Chg #68: Add support for wildcard repo match

## Details

 * Bugfix #67: Fixed typecasts within billing API

   In some cases it happened that the GitHub billing API responded with floats, but we only
   accepted integers. With this fix any number should be casted to floats all the time.

   https://github.com/promhippie/github_exporter/issues/67

 * Change #45: Change docker image name

   We should use the same docker image name as all the other exporters within this organization. So
   we renamed the image from `promhippie/github_exporter` to `promhippie/github-exporter`
   to have the same naming convention as for the other exporters.

   https://github.com/promhippie/github_exporter/issues/45

 * Change #42: Drop darwin/386 release builds

   We dropped the build of 386 builds on Darwin as this architecture is not supported by current Go
   versions anymore.

   https://github.com/promhippie/github_exporter/issues/42

 * Change #46: Generate metrics documentation

   We have added a script to automatically generate the available metrics within the
   documentation to prevent any documentation gaps. Within the `hack/` folder we got a small Go
   script which parses the available collectors and updates the documentation partial based on
   that.

   https://github.com/promhippie/github_exporter/issues/46

 * Change #71: Integrate standard web config

   We integrated the new web config from the Prometheus toolkit which provides a configuration
   for TLS support and also some basic builtin authentication. For the detailed configuration
   you check out the documentation.

   https://github.com/promhippie/github_exporter/issues/71

 * Change #68: Add support for wildcard repo match

   We integrated the functionality to add a wildcard matching for repository names to export
   metrics from. Now you don't need to add every single repo you want to match, you can add a whole
   organization.

   https://github.com/promhippie/github_exporter/issues/68


# Changelog for 1.0.1

The following sections list the changes for 1.0.1.

## Summary

 * Fix #49: Fixed pointer references within exporters

## Details

 * Bugfix #49: Fixed pointer references within exporters

   So far we directly accessed common attributes on repos within the collectors, but as there
   could be corner cases where some attribute could be missing we add conditions to make sure only
   set values are getting called.

   https://github.com/promhippie/github_exporter/issues/49


# Changelog for 1.0.0

The following sections list the changes for 1.0.0.

## Summary

 * Fix #18: Handle private repos counts not available to non-organization member
 * Chg #12: Refactor structure and integrate more metrics
 * Enh #19: Allow insecure TLS via CLI flag
 * Enh #20: Add flag to set /metrics endpoint request timeout

## Details

 * Bugfix #18: Handle private repos counts not available to non-organization member

   Fix the case where the account used to query GitHub API is not a member of a given organisation, it
   resulted in a segfault.

   https://github.com/promhippie/github_exporter/pull/18

 * Change #12: Refactor structure and integrate more metrics

   The structure of the repository should get overhauled and beside that we should gather more
   metrics per organization and per repository. Additionally we should also add support for
   GitHub Enterprise as requested within
   [#10](https://github.com/promhippie/github_exporter/issues/10).

   https://github.com/promhippie/github_exporter/issues/12

 * Enhancement #19: Allow insecure TLS via CLI flag

   In some cases it can be desirable to ignore certificate errors from the GitHub API - such as in the
   case of connecting to a private instance of GitHub Enterprise which uses a self-signed cert.
   This is exposed via the environment variable `GITHUB_EXPORTER_TLS_INSECURE` and the flag
   `--github.insecure`.

   https://github.com/promhippie/github_exporter/pull/19

 * Enhancement #20: Add flag to set /metrics endpoint request timeout

   When pulling a lot of data from the GitHub API, in some cases the default 10s timeout on the
   `/metrics` endpoint can be insufficient. This option allows the timeout to be configured via
   `GITHUB_EXPORTER_WEB_TIMEOUT` or `--web.timeout`

   https://github.com/promhippie/github_exporter/pull/20


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


