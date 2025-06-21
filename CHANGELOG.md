# Changelog for 5.0.0

The following sections list the changes for 5.0.0.

## Summary

 * BREAKING: Complete restructure of billing metrics for Enhanced Billing Platform support
 * BREAKING: Authentication requirements changed for enterprise billing endpoints
 * Enhancement: New dimensional billing metrics with repository attribution and cost breakdown
 * Enhancement: Configurable query parameters for time range and cost center filtering
 * Enhancement: Granularity controls for managing Prometheus cardinality and performance

## Details

 * BREAKING CHANGE: Complete restructure of billing metrics for Enhanced Billing Platform support

   GitHub has migrated organizations to the Enhanced Billing Platform, causing legacy billing endpoints to return 410 errors. This release completely restructures billing metrics to use the new unified billing API endpoints.

   **Legacy metrics removed (10+ metrics):**
   - `github_action_billing_*` (4 metrics)
   - `github_package_billing_*` (3 metrics) 
   - `github_storage_billing_*` (3 metrics)

   **New unified metrics (5 metrics with rich dimensional labels):**
   - `github_billing_usage_quantity` - Usage quantities with full dimensional breakdown
   - `github_billing_usage_cost_gross` - Gross costs before discounts
   - `github_billing_usage_cost_net` - Net costs after discounts (actual charges)
   - `github_billing_usage_cost_discount` - Discount amounts applied
   - `github_billing_usage_price_per_unit` - Per-unit pricing information

   **New label dimensions:**
   - `type` - "org" or "enterprise"
   - `name` - Organization/enterprise name
   - `product` - "Actions", "Packages", etc.
   - `sku` - Product-specific SKU (e.g., "Actions Linux")
   - `unit_type` - "minutes", "gigabytes", "GigabyteHours"
   - `date` - Usage date (YYYY-MM-DD)
   - `organization` - Organization name from API
   - `repository` - Repository name (cardinality limited)

   **Migration required:** All existing dashboards, alerts, and queries using billing metrics must be updated. See migration guide for conversion examples.

 * BREAKING CHANGE: Authentication requirements changed for enterprise billing endpoints

   Enterprise billing endpoints now require Personal Access Tokens (Classic) only. Fine-grained tokens are not supported for enterprise billing.

   **Organizations** (`/organizations/{org}/settings/billing/usage`):
   - Fine-grained PATs: Supported with "Administration" org permissions (read)
   - Classic PATs: Supported

   **Enterprises** (`/enterprises/{enterprise}/settings/billing/usage`):
   - Fine-grained PATs: NOT supported
   - Classic PATs: Required with "manage_billing:enterprise" scope

   **Migration required:** Enterprise users with fine-grained tokens must switch to Classic PATs.

 * Enhancement: New dimensional billing metrics with repository attribution and cost breakdown

   The new metrics provide unprecedented granularity for GitHub billing analysis:
   - Daily usage granularity for time-series analysis
   - Repository-level attribution to identify high-usage repos
   - Cost breakdown with gross, net, and discount amounts
   - Product/SKU specificity for detailed analysis
   - Better Prometheus querying capabilities

   Repository cardinality is limited to 100 repositories per entity (with excess aggregated as "other") to prevent Prometheus performance issues.

 * Enhancement: Configurable query parameters for time range and cost center filtering

   New configuration options allow precise control over data collection scope:

   ```yaml
   target:
     billing:
       year: 2024                    # Filter by specific year
       month: 12                     # Filter by specific month
       day: 15                       # Filter by specific day
       hour: 14                      # Filter by specific hour
       cost_center_id: "engineering" # Enterprise cost center filtering
   ```

   Query parameters include comprehensive validation and default to current month to prevent overwhelming data downloads.

 * Enhancement: Granularity controls for managing Prometheus cardinality and performance

   New configuration options provide fine-grained control over metric dimensions to balance detail with performance:

   ```yaml
   target:
     billing:
       disable_repository_labels: false   # Control repository attribution
       disable_date_labels: false        # Control temporal granularity
       disable_organization_labels: false # Control organization labels
       enabled_metrics:                   # Select specific metrics
         - "quantity"
         - "cost_net"
       max_repositories: 50              # Override repository limit
   ```

   These controls allow users to optimize for their specific monitoring needs and Prometheus resource constraints.

**Breaking Change Notice:** This is a major version release with significant breaking changes. All users collecting billing metrics must update their configuration, dashboards, and alerting rules. See the migration guide for detailed conversion examples and best practices.

# Changelog for 4.1.1

The following sections list the changes for 4.1.1.

## Summary

 * Fix #485: Missing field updates for workflow jobs

## Details

 * Bugfix #485: Missing field updates for workflow jobs

   With the previous behavior the runner specific labels for worklow job metrics
   have always been empty. They we always set after the job started but have never
   been part of the database table update process.

   https://github.com/promhippie/github_exporter/issues/485


# Changelog for 4.1.0

The following sections list the changes for 4.1.0.

## Summary

 * Enh #443: Extra config option to change the database purge window
 * Enh #469: Optional conclusion label for workflow jobs

## Details

 * Enhancement #443: Extra config option to change the database purge window

   We added a new option to optionally specify a different purge window for the
   database content if you are directly using a database source to read the metrics
   details.

   https://github.com/promhippie/github_exporter/pull/443

 * Enhancement #469: Optional conclusion label for workflow jobs

   We added another optional label which includes the `conclusion` value as an
   extra label as this could be different than the `status` value. This label got
   to be enabled by the `--collector.workflow_jobs.labels` flag or the
   `GITHUB_EXPORTER_WORKFLOW_JOBS_LABELS` environment variable.

   https://github.com/promhippie/github_exporter/pull/469


# Changelog for 4.0.1

The following sections list the changes for 4.0.1.

## Summary

 * Fix #413: Show errors for database ping
 * Fix #414: Fix PostgreSQL identifiers to be bigint
 * Fix #417: Runner ID was being set to run ID for workflow_job
 * Fix #418: Include identifier in labels for workflow jobs
 * Fix #418: Update conclusion and completed_at columns

## Details

 * Bugfix #413: Show errors for database ping

   To get more details why a ping to the database fails while starting the exporter
   we have added an error output to the logging.

   https://github.com/promhippie/github_exporter/pull/413

 * Bugfix #414: Fix PostgreSQL identifiers to be bigint

   Since Github submit 64bit integers for the identifier of workflow jobs we had to
   fix the type to `BIGINT` for the database schema to avoid errors related to
   store events.

   https://github.com/promhippie/github_exporter/pull/414

 * Bugfix #417: Runner ID was being set to run ID for workflow_job

   While implementing the workflow job collector we simply attached the wrong
   identifier to the runner id which was base on the run ID. Future webhooks will
   store the right ID now.

   https://github.com/promhippie/github_exporter/pull/417

 * Bugfix #418: Include identifier in labels for workflow jobs

   To avoid errors related to already scraped metrics we have added the identifier
   to the default labels for the workflow job collector.

   https://github.com/promhippie/github_exporter/pull/418

 * Bugfix #418: Update conclusion and completed_at columns

   For the new workflow job collector we had been missing the `conclusion` and
   `completed_at` values, they will be stored by future webhook events.

   https://github.com/promhippie/github_exporter/pull/418


# Changelog for 4.0.0

The following sections list the changes for 4.0.0.

## Summary

 * Chg #393: Switch to official logging library
 * Chg #405: New metrics and configs for workflow job collector
 * Chg #412: Change config and metric names for workflows
 * Enh #368: Add `actor.login` label to workflow collector

## Details

 * Change #393: Switch to official logging library

   Since there have been a structured logger part of the Go standard library we
   thought it's time to replace the library with that. Be aware that log messages
   should change a little bit.

   https://github.com/promhippie/github_exporter/issues/393

 * Change #405: New metrics and configs for workflow job collector

   We have added a new metric for the duration of the time since the workflow job
   was created, defined in minutes. Beside that we have added two additional
   configurations to query the workflows for a specific status and you are able to
   define a different timeframe than 12 hours now.

   https://github.com/promhippie/github_exporter/pull/405

 * Change #412: Change config and metric names for workflows

   We introduced a BREAKING CHANGE by renaming the config variables and metrics
   related to the workflows. Previously you had to set `--collector.workflows` or
   `GITHUB_EXPORTER_COLLECTOR_WORKFLOWS` to enable the collector, for being
   consistent with the new workflow job collector we have renamed them to
   `--collector.workflow_runs` and `GITHUB_EXPORTER_COLLECTOR_WORKFLOW_JOBS`, so be
   aware about that. Additionally we have also renamed the metrics a tiny bit
   matching the same suffix. We renamed them as an example from
   `github_workflow_status` to `github_workflow_run_status`.

   https://github.com/promhippie/github_exporter/pull/412

 * Enhancement #368: Add `actor.login` label to workflow collector

   Export an `actor.login` label for workflow collector which could be added to the
   labels optionally, please keep the label cardinality in mind while enabling this
   label.

   https://github.com/promhippie/github_exporter/issues/368


# Changelog for 3.1.2

The following sections list the changes for 3.1.2.

## Summary

 * Fix #296: Drop out of order event if same timestamp as existing completed event

## Details

 * Bugfix #296: Drop out of order event if same timestamp as existing completed event

   We've fixed the behavior of out of order workflow events with the same
   timestamp, by preferring an already recorded `completed` event

   https://github.com/promhippie/github_exporter/issues/296
   https://github.com/promhippie/github_exporter/pull/298
   https://github.com/promhippie/github_exporter/pull/312


# Changelog for 3.1.1

The following sections list the changes for 3.1.1.

## Summary

 * Fix #296: Drop out of order workflow webhook events
 * Fix #300: Rename package followin upstream genji package

## Details

 * Bugfix #296: Drop out of order workflow webhook events

   We've fixed the behavior of out of order workflow events by dropping them, since
   the events order is not guaranteed by GitHub

   https://github.com/promhippie/github_exporter/issues/296
   https://github.com/promhippie/github_exporter/pull/298

 * Bugfix #300: Rename package followin upstream genji package

   The upstream project of the genji database implementation have been renamed, to
   stay being able to upgrade the library we are force to rename it within this
   project as well. We keep the driver name compatible to existing installations.

   https://github.com/promhippie/github_exporter/issues/300


# Changelog for 3.1.0

The following sections list the changes for 3.1.0.

## Summary

 * Fix #278: Create SQLite directory if it doesn't exist
 * Enh #277: Configurable labels for runner metrics
 * Enh #281: Add metrics for workflow timestamps

## Details

 * Bugfix #278: Create SQLite directory if it doesn't exist

   We have integrated a fix to always try to create the directory for the SQLite
   database if it doesn't exist. The exporter will fail to start if the directory
   does not exist and if it fails to create the directory for the database file.

   https://github.com/promhippie/github_exporter/issues/278

 * Enhancement #277: Configurable labels for runner metrics

   Initially we had a static list of available labels for the runner metrics, with
   this change we are getting more labels like the GitHub runner labels as
   Prometheus labels while they are configurable to avoid a high cardinality of
   Prometheus labels. Now you are able to get labels for `owner`, `id`, `name`,
   `os`, `status` and optionally `labels`.

   https://github.com/promhippie/github_exporter/issues/277

 * Enhancement #281: Add metrics for workflow timestamps

   We added new metrics to show multiple timestamps for the workflows like when the
   workflow have been created, updated and started. Please look at the
   documentation for the exact naming of these new metrics.

   https://github.com/promhippie/github_exporter/issues/281


# Changelog for 3.0.1

The following sections list the changes for 3.0.1.

## Summary

 * Fix #270: Correctly store and retrieve records
 * Fix #272: Add SQLite and Genji if supported

## Details

 * Bugfix #270: Correctly store and retrieve records

   We had introduced a bug while switching between golang's sqlx and sql packages,
   with this fix all workflows should be stored and retrieved correctly.

   You got to make sure to delete the database which have been created with the
   3.0.0 release as the migration setup have been changed.

   https://github.com/promhippie/github_exporter/issues/270

 * Bugfix #272: Add SQLite and Genji if supported

   We haven't been able to build all supported binaries as some have been lacking
   support for the use libraries for SQLite and Genji. We have added build tags
   which enables or disables the database drivers if needed.

   https://github.com/promhippie/github_exporter/pull/272


# Changelog for 3.0.0

The following sections list the changes for 3.0.0.

## Summary

 * Fix #267: Use right date for workflow durations
 * Chg #245: Read secrets form files
 * Enh #261: Rebuild workflow collector based on webhooks

## Details

 * Bugfix #267: Use right date for workflow durations

   We had used the creation date to detect the workflow duration, this have been
   replaced by the run started date to show the right duration time for the
   workflow.

   https://github.com/promhippie/github_exporter/issues/267

 * Change #245: Read secrets form files

   We have added proper support to load secrets like the token or the private key
   for app authentication from files or from base64-encoded strings. Just provide
   the flags or environment variables for token or private key with a DSN formatted
   string like `file://path/to/file` or `base64://Zm9vYmFy`.

   Since the private key for GitHub App authentication had been provided in
   base64-encoded format this is a breaking change as this won't work anymore until
   you prefix the value with `base64://`.

   https://github.com/promhippie/github_exporter/pull/245

 * Enhancement #261: Rebuild workflow collector based on webhooks

   We have rebuilt the workflow collector based on GitHub webhooks to get a more
   stable behavior and to avoid running into many rate limits. Receiving changes
   directly from gitHub instead of requesting it should improve the behavior a lot.
   Please read the docs to see how you can setup the required webhooks within
   GitHub, opening up the exporter for GitHub should be the most simple part.

   https://github.com/promhippie/github_exporter/issues/261


# Changelog for 2.4.0

The following sections list the changes for 2.4.0.

## Summary

 * Enh #247: Add metrics for organization seats

## Details

 * Enhancement #247: Add metrics for organization seats

   We had been missing metrics about available and remaining seats within
   organizations. This change adds the two metrics `github_org_filled_seats` and
   `github_org_seats` related to that.

   https://github.com/promhippie/github_exporter/pull/247


# Changelog for 2.3.0

The following sections list the changes for 2.3.0.

## Summary

 * Fix #230: Properly handle workflow status results
 * Enh #228: Enable app support for GitHub Enterprise

## Details

 * Bugfix #230: Properly handle workflow status results

   While a workflow is running the conclusion property provides an empty string, in
   order to get a proper value in this case we have changed the metric value to the
   status property which offers a useable fallback.

   https://github.com/promhippie/github_exporter/issues/230

 * Enhancement #228: Enable app support for GitHub Enterprise

   Previously we had support for GitHub applications for the SaaS version only,
   with this change you are also able to use application registration for GitHub
   Enterprise.

   https://github.com/promhippie/github_exporter/issues/228


# Changelog for 2.2.1

The following sections list the changes for 2.2.1.

## Summary

 * Fix #216: Resolve nil pointer issues with responses

## Details

 * Bugfix #216: Resolve nil pointer issues with responses

   We have introduced previously a feature where we made sure to close all response
   bodies, but we forgot to properly check if these responses are really valid
   which lead to nil pointer dereference issues.

   https://github.com/promhippie/github_exporter/issues/216


# Changelog for 2.2.0

The following sections list the changes for 2.2.0.

## Summary

 * Fix #190: Prevent concurrent scrapes
 * Enh #193: Integrate option pprof profiling
 * Enh #200: New metrics and configs for workflow collector

## Details

 * Bugfix #190: Prevent concurrent scrapes

   If the exporter got some kind of duplicated repository names configured it lead
   to errors because the label combination had been scraped already. We have added
   some simple checks to prevent duplicated exports an all currently available
   collectors.

   https://github.com/promhippie/github_exporter/issues/190

 * Enhancement #193: Integrate option pprof profiling

   We have added an option to enable a pprof endpoint for proper profiling support
   with the help of tools like Parca. The endpoint `/debug/pprof` can now
   optionally be enabled to get the profiling details for catching potential memory
   leaks.

   https://github.com/promhippie/github_exporter/pull/193

 * Enhancement #200: New metrics and configs for workflow collector

   We have added a new metric for the duration of the time since the workflow run
   was created, defined in minutes. Beside that we have added two additional
   configurations to query the workflows for a specific status and you are able to
   define a different timeframe than 12 hours now. Finally we have also added the
   run ID to the labels.

   https://github.com/promhippie/github_exporter/pull/200
   https://github.com/promhippie/github_exporter/pull/214


# Changelog for 2.1.0

The following sections list the changes for 2.1.0.

## Summary

 * Enh #187: Integrate a flage to define pagination size

## Details

 * Enhancement #187: Integrate a flage to define pagination size

   There had been no way to set a specific page size before, we have added an
   option/flag that you are able to define the pagination size on your own to avoid
   running into rate limits.

   https://github.com/promhippie/github_exporter/issues/187


# Changelog for 2.0.1

The following sections list the changes for 2.0.1.

## Summary

 * Fix #185: Improve parsing of private key for GitHub app

## Details

 * Bugfix #185: Improve parsing of private key for GitHub app

   Previously we always checked if a file with the value exists if a private key
   for GitHub app authentication have been provided, now I have switched it to try
   to base64 decode the string first, and try to load the file afterwards which
   works more reliable and avoids leaking the private key into the log output.

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

   We introduced metrics for GitHub self-hosted runners but we missed some
   important labels as remaining todos. With this change this gets corrected to
   properly show the repo/org/enterprise where the runner have been attached to.

   https://github.com/promhippie/github_exporter/pull/184

 * Enhancement #123: Add metrics for GitHub runners

   We've added new metrics for selfhosted runners used per repo, org or enterprise
   to give the ability to check if the runners are online and busy.

   https://github.com/promhippie/github_exporter/issues/123

 * Enhancement #123: Add metrics for GitHub workflows

   We've added new metrics for the observability of GitHub workflows/actions, they
   are disabled by default because it could result in a high cardinality of the
   labels.

   https://github.com/promhippie/github_exporter/issues/123

 * Enhancement #174: Merge all billing related metrics

   We've merged the three available collectors only related to billing into a
   single billing collector to reduce the required options and simply because they
   related to each other.

   https://github.com/promhippie/github_exporter/pull/174

 * Enhancement #174: Update all releated dependencies

   We've updated all dependencies to the latest available versions, including the
   build tools provided by Bingo.

   https://github.com/promhippie/github_exporter/pull/174

 * Enhancement #183: Integrate admin stats for GitHub enterprise

   We've integrated another collector within this exporter to provide admin stats
   as metrics to get a general overview about the amount of repos, issues, pull
   requests and so on. Special thanks for the great initial work by @mafrosis, your
   effort is highly appreciated.

   https://github.com/promhippie/github_exporter/issues/183
   https://github.com/promhippie/github_exporter/pull/23

 * Enhancement #184: Use getter functions to get values

   To reduce the used boilerplate code and to better use the GitHub library we have
   updated most of the available collectors to simply use the provided getter
   functions instead of checking for nil values everywhere on our own.

   https://github.com/promhippie/github_exporter/pull/184


# Changelog for 1.2.0

The following sections list the changes for 1.2.0.

## Summary

 * Enh #145: Add support for GitHub app

## Details

 * Enhancement #145: Add support for GitHub app

   We've added an integration with GitHub Apps. Organizations which are not able to
   use tokens for the API access are able to register a GitHub app for the API
   requests of this exporter now.

   https://github.com/promhippie/github_exporter/pull/145


# Changelog for 1.1.1

The following sections list the changes for 1.1.1.

## Summary

 * Fix #75: Fixed wildcard matching for private repos

## Details

 * Bugfix #75: Fixed wildcard matching for private repos

   We have fixed the ability to match by globbing/wildcard for private
   repositories. So far the private repositories have been missing with the latest
   globbing/wildcard matching.

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

   In some cases it happened that the GitHub billing API responded with floats, but
   we only accepted integers. With this fix any number should be casted to floats
   all the time.

   https://github.com/promhippie/github_exporter/issues/67

 * Change #42: Drop darwin/386 release builds

   We dropped the build of 386 builds on Darwin as this architecture is not
   supported by current Go versions anymore.

   https://github.com/promhippie/github_exporter/issues/42

 * Change #45: Change docker image name

   We should use the same docker image name as all the other exporters within this
   organization. So we renamed the image from `promhippie/github_exporter` to
   `promhippie/github-exporter` to have the same naming convention as for the other
   exporters.

   https://github.com/promhippie/github_exporter/issues/45

 * Change #46: Generate metrics documentation

   We have added a script to automatically generate the available metrics within
   the documentation to prevent any documentation gaps. Within the `hack/` folder
   we got a small Go script which parses the available collectors and updates the
   documentation partial based on that.

   https://github.com/promhippie/github_exporter/issues/46

 * Change #68: Add support for wildcard repo match

   We integrated the functionality to add a wildcard matching for repository names
   to export metrics from. Now you don't need to add every single repo you want to
   match, you can add a whole organization.

   https://github.com/promhippie/github_exporter/issues/68

 * Change #71: Integrate standard web config

   We integrated the new web config from the Prometheus toolkit which provides a
   configuration for TLS support and also some basic builtin authentication. For
   the detailed configuration you check out the documentation.

   https://github.com/promhippie/github_exporter/issues/71


# Changelog for 1.0.1

The following sections list the changes for 1.0.1.

## Summary

 * Fix #49: Fixed pointer references within exporters

## Details

 * Bugfix #49: Fixed pointer references within exporters

   So far we directly accessed common attributes on repos within the collectors,
   but as there could be corner cases where some attribute could be missing we add
   conditions to make sure only set values are getting called.

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

   Fix the case where the account used to query GitHub API is not a member of a
   given organisation, it resulted in a segfault.

   https://github.com/promhippie/github_exporter/pull/18

 * Change #12: Refactor structure and integrate more metrics

   The structure of the repository should get overhauled and beside that we should
   gather more metrics per organization and per repository. Additionally we should
   also add support for GitHub Enterprise as requested within
   [#10](https://github.com/promhippie/github_exporter/issues/10).

   https://github.com/promhippie/github_exporter/issues/12

 * Enhancement #19: Allow insecure TLS via CLI flag

   In some cases it can be desirable to ignore certificate errors from the GitHub
   API - such as in the case of connecting to a private instance of GitHub
   Enterprise which uses a self-signed cert. This is exposed via the environment
   variable `GITHUB_EXPORTER_TLS_INSECURE` and the flag `--github.insecure`.

   https://github.com/promhippie/github_exporter/pull/19

 * Enhancement #20: Add flag to set /metrics endpoint request timeout

   When pulling a lot of data from the GitHub API, in some cases the default 10s
   timeout on the `/metrics` endpoint can be insufficient. This option allows the
   timeout to be configured via `GITHUB_EXPORTER_WEB_TIMEOUT` or `--web.timeout`

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

   The previous metric `github_valid_response` doesn't match the Prometheus
   conventions, so it have been renamed to `github_up` which properly signals if
   the exporter can gather metrics or not.

   https://github.com/promhippie/github_exporter/issues/2

 * Change #4: Enforce a repo or an org flag

   The exporter requires at least one organization or repository to work properly,
   integrated a check that something have been set when launching the exporter.

   https://github.com/promhippie/github_exporter/issues/4


