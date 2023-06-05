# Changelog for unreleased

The following sections list the changes for unreleased.

## Summary

 * Fix #190: Prevent concurrent scrapes
 * Enh #193: Integrate option pprof profiling
 * Enh #200: New metrics and configurations for the workflow collector

## Details

 * Bugfix #190: Prevent concurrent scrapes

   If the exporter got some kind of duplicated repository names configured it lead to errors
   because the label combination had been scraped already. We have added some simple checks to
   prevent duplicated exports an all currently available collectors.

   https://github.com/promhippie/github_exporter/issues/190

 * Enhancement #193: Integrate option pprof profiling

   We have added an option to enable a pprof endpoint for proper profiling support with the help of
   tools like Parca. The endpoint `/debug/pprof` can now optionally be enabled to get the
   profiling details for catching potential memory leaks.

   https://github.com/promhippie/github_exporter/pull/193

 * Enhancement #200: New metrics and configurations for the workflow collector

   1. Added a new metric for the duration, in minutes, of the time since the run was created:
   github_workflow_duration_run_created_minutes 2. Added 2 optional configuration options
   for the workflows exporter: 1. Query workflows with a specific status (Default to any) 2. Set
   the window history (defaults to 12 hours) 3. Added run_id label

   https://github.com/promhippie/github_exporter/pull/200


# Changelog for 2.1.0

The following sections list the changes for 2.1.0.

## Summary

 * Enh #187: Integrate a flage to define pagination size

## Details

 * Enhancement #187: Integrate a flage to define pagination size

   There had been no way to set a specific page size before, we have added an option/flag that you are
   able to define the pagination size on your own to avoid running into rate limits.

   https://github.com/promhippie/github_exporter/issues/187


# Changelog for 2.0.1

The following sections list the changes for 2.0.1.

## Summary

 * Fix #185: Improve parsing of private key for GitHub app

## Details

 * Bugfix #185: Improve parsing of private key for GitHub app

   Previously we always checked if a file with the value exists if a private key for GitHub app
   authentication have been provided, now I have switched it to try to base64 decode the string
   first, and try to load the file afterwards which works more reliable and avoids leaking the
   private key into the log output.

   https://github.com/promhippie/github_exporter/issues/185


# Changelog for 2.0.0

The following sections list the changes for 2.0.0.

## Summary

 * Fix #184: Set right name/owner labels for runner metrics
 * Enh #123: Add metrics for GitHub runners
 * Enh #123: Add metrics for GitHub workflows
 * Enh #174: Merge all billing related metrics
 * Enh #174: Update all releated dependencies
 * Enh #183: Integrate admin stats for GitHub enterprise
 * Enh #184: Use getter functions to get values

## Details

 * Bugfix #184: Set right name/owner labels for runner metrics

   We introduced metrics for GitHub self-hosted runners but we missed some important labels as
   remaining todos. With this change this gets corrected to properly show the
   repo/org/enterprise where the runner have been attached to.

   https://github.com/promhippie/github_exporter/pull/184

 * Enhancement #123: Add metrics for GitHub runners

   We've added new metrics for selfhosted runners used per repo, org or enterprise to give the
   ability to check if the runners are online and busy.

   https://github.com/promhippie/github_exporter/issues/123

 * Enhancement #123: Add metrics for GitHub workflows

   We've added new metrics for the observability of GitHub workflows/actions, they are disabled
   by default because it could result in a high cardinality of the labels.

   https://github.com/promhippie/github_exporter/issues/123

 * Enhancement #174: Merge all billing related metrics

   We've merged the three available collectors only related to billing into a single billing
   collector to reduce the required options and simply because they related to each other.

   https://github.com/promhippie/github_exporter/pull/174

 * Enhancement #174: Update all releated dependencies

   We've updated all dependencies to the latest available versions, including the build tools
   provided by Bingo.

   https://github.com/promhippie/github_exporter/pull/174

 * Enhancement #183: Integrate admin stats for GitHub enterprise

   We've integrated another collector within this exporter to provide admin stats as metrics to
   get a general overview about the amount of repos, issues, pull requests and so on. Special
   thanks for the great initial work by @mafrosis, your effort is highly appreciated.

   https://github.com/promhippie/github_exporter/issues/183
   https://github.com/promhippie/github_exporter/pull/23

 * Enhancement #184: Use getter functions to get values

   To reduce the used boilerplate code and to better use the GitHub library we have updated most of
   the available collectors to simply use the provided getter functions instead of checking for
   nil values everywhere on our own.

   https://github.com/promhippie/github_exporter/pull/184


# Changelog for 1.2.0

The following sections list the changes for 1.2.0.

## Summary

 * Enh #145: Add support for GitHub app

## Details

 * Enhancement #145: Add support for GitHub app

   We've added an integration with GitHub Apps. Organizations which are not able to use tokens for
   the API access are able to register a GitHub app for the API requests of this exporter now.

   https://github.com/promhippie/github_exporter/pull/145


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
 * Chg #42: Drop darwin/386 release builds
 * Chg #45: Change docker image name
 * Chg #46: Generate metrics documentation
 * Chg #68: Add support for wildcard repo match
 * Chg #71: Integrate standard web config

## Details

 * Bugfix #67: Fixed typecasts within billing API

   In some cases it happened that the GitHub billing API responded with floats, but we only
   accepted integers. With this fix any number should be casted to floats all the time.

   https://github.com/promhippie/github_exporter/issues/67

 * Change #42: Drop darwin/386 release builds

   We dropped the build of 386 builds on Darwin as this architecture is not supported by current Go
   versions anymore.

   https://github.com/promhippie/github_exporter/issues/42

 * Change #45: Change docker image name

   We should use the same docker image name as all the other exporters within this organization. So
   we renamed the image from `promhippie/github_exporter` to `promhippie/github-exporter`
   to have the same naming convention as for the other exporters.

   https://github.com/promhippie/github_exporter/issues/45

 * Change #46: Generate metrics documentation

   We have added a script to automatically generate the available metrics within the
   documentation to prevent any documentation gaps. Within the `hack/` folder we got a small Go
   script which parses the available collectors and updates the documentation partial based on
   that.

   https://github.com/promhippie/github_exporter/issues/46

 * Change #68: Add support for wildcard repo match

   We integrated the functionality to add a wildcard matching for repository names to export
   metrics from. Now you don't need to add every single repo you want to match, you can add a whole
   organization.

   https://github.com/promhippie/github_exporter/issues/68

 * Change #71: Integrate standard web config

   We integrated the new web config from the Prometheus toolkit which provides a configuration
   for TLS support and also some basic builtin authentication. For the detailed configuration
   you check out the documentation.

   https://github.com/promhippie/github_exporter/issues/71


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


# Changelog for 0.1.0

The following sections list the changes for 0.1.0.

## Summary

 * Chg #11: Initial release of basic version

## Details

 * Change #11: Initial release of basic version

   Just prepared an initial basic version which could be released to the public.

   https://github.com/promhippie/github_exporter/issues/11


# Changelog for 0.2.0

The following sections list the changes for 0.2.0.

## Summary

 * Chg #2: Renamed valid_response metric to up metric
 * Chg #4: Enforce a repo or an org flag

## Details

 * Change #2: Renamed valid_response metric to up metric

   The previous metric `github_valid_response` doesn't match the Prometheus conventions, so
   it have been renamed to `github_up` which properly signals if the exporter can gather metrics
   or not.

   https://github.com/promhippie/github_exporter/issues/2

 * Change #4: Enforce a repo or an org flag

   The exporter requires at least one organization or repository to work properly, integrated a
   check that something have been set when launching the exporter.

   https://github.com/promhippie/github_exporter/issues/4


