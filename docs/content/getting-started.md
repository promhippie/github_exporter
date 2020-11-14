---
title: "Getting Started"
date: 2020-10-28T00:00:00+00:00
anchor: "getting-started"
weight: 10
---

## Installation

We won't cover further details how to properly setup [Prometheus](https://prometheus.io) itself, we will only cover some basic setup based on [docker-compose](https://docs.docker.com/compose/). But if you want to run this exporter without [docker-compose](https://docs.docker.com/compose/) you should be able to adopt that to your needs.

First of all we need to prepare a configuration for [Prometheus](https://prometheus.io) that includes the exporter as a target based on a static host mapping which is just the [docker-compose](https://docs.docker.com/compose/) container name, e.g. `github_exporter`.

{{< highlight yaml >}}
global:
  scrape_interval: 1m
  scrape_timeout: 10s
  evaluation_interval: 1m

scrape_configs:
- job_name: github
  static_configs:
  - targets:
    - github_exporter:9504
{{< / highlight >}}

After preparing the configuration we need to create the `docker-compose.yml` within the same folder, this `docker-compose.yml` starts a simple [Prometheus](https://prometheus.io) instance together with the exporter. Don't forget to update the exporter envrionment variables with the required credentials.

{{< highlight yaml >}}
version: '2'

volumes:
  prometheus:

services:
  prometheus:
    image: prom/prometheus:latest
    restart: always
    ports:
      - 9090:9090
    volumes:
      - prometheus:/prometheus
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  github_exporter:
    image: promhippie/github_exporter:latest
    restart: always
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

Since our `latest` Docker tag always refers to the `master` branch of the Git repository you should always use some fixed version. You can see all available tags at our [DockerHub repository](https://hub.docker.com/r/promhippie/github_exporter/tags/), there you will see that we also provide a manifest, you can easily start the exporter on various architectures without any change to the image name. You should apply a change like this to the `docker-compose.yml`:

{{< highlight diff >}}
  hcloud-exporter:
-   image: promhippie/github_exporter:latest
+   image: promhippie/github_exporter:1.0.0
    restart: always
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

If you want to access the exporter directly you should bind it to a local port, otherwise only [Prometheus](https://prometheus.io) will have access to the exporter. For debugging purpose or just to discover all available metrics directly you can apply this change to your `docker-compose.yml`, after that you can access it directly at [http://localhost:9504/metrics](http://localhost:9504/metrics):

{{< highlight diff >}}
  hcloud-exporter:
    image: promhippie/github_exporter:latest
    restart: always
+   ports:
+     - 127.0.0.1:9504:9504
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

Finally the exporter should be configured fine, let's start this stack with [docker-compose](https://docs.docker.com/compose/), you just need to execute `docker-compose up` within the directory where you have stored the `prometheus.yml` and `docker-compose.yml`.

{{< highlight txt >}}
Creating network "example_default" with the default driver
Creating volume "example_prometheus" with default driver
Creating example_github_exporter_1 ... done
Creating example_prometheus_1      ... done
Attaching to example_github_exporter_1, example_prometheus_1
github_exporter_1  | level=info ts=2020-10-29T07:15:08.8972594Z msg="Launching GitHub Exporter" version=9f39482 revision=9f39482 date=20201028 go=go1.14.2
github_exporter_1  | level=info ts=2020-10-29T07:15:08.8976418Z msg="Starting metrics server" addr=0.0.0.0:9504
prometheus_1       | level=info ts=2020-10-29T07:15:09.198Z caller=main.go:315 msg="No time or size retention was set so using the default time retention" duration=15d
prometheus_1       | level=info ts=2020-10-29T07:15:09.198Z caller=main.go:353 msg="Starting Prometheus" version="(version=2.22.0, branch=HEAD, revision=0a7fdd3b76960808c3a91d92267c3d815c1bc354)"
prometheus_1       | level=info ts=2020-10-29T07:15:09.198Z caller=main.go:358 build_context="(go=go1.15.3, user=root@6321101b2c50, date=20201015-12:29:59)"
prometheus_1       | level=info ts=2020-10-29T07:15:09.198Z caller=main.go:359 host_details="(Linux 4.19.76-linuxkit #1 SMP Tue May 26 11:42:35 UTC 2020 x86_64 95702f475360 (none))"
prometheus_1       | level=info ts=2020-10-29T07:15:09.198Z caller=main.go:360 fd_limits="(soft=1048576, hard=1048576)"
prometheus_1       | level=info ts=2020-10-29T07:15:09.199Z caller=main.go:361 vm_limits="(soft=unlimited, hard=unlimited)"
prometheus_1       | level=info ts=2020-10-29T07:15:09.204Z caller=web.go:516 component=web msg="Start listening for connections" address=0.0.0.0:9090
prometheus_1       | level=info ts=2020-10-29T07:15:09.204Z caller=main.go:712 msg="Starting TSDB ..."
prometheus_1       | level=info ts=2020-10-29T07:15:09.208Z caller=head.go:642 component=tsdb msg="Replaying on-disk memory mappable chunks if any"
prometheus_1       | level=info ts=2020-10-29T07:15:09.208Z caller=head.go:656 component=tsdb msg="On-disk memory mappable chunks replay completed" duration=9.5µs
prometheus_1       | level=info ts=2020-10-29T07:15:09.208Z caller=head.go:662 component=tsdb msg="Replaying WAL, this may take a while"
prometheus_1       | level=info ts=2020-10-29T07:15:09.209Z caller=head.go:714 component=tsdb msg="WAL segment loaded" segment=0 maxSegment=0
prometheus_1       | level=info ts=2020-10-29T07:15:09.209Z caller=head.go:719 component=tsdb msg="WAL replay completed" checkpoint_replay_duration=39.9µs wal_replay_duration=803µs total_replay_duration=923.1µs
prometheus_1       | level=info ts=2020-10-29T07:15:09.210Z caller=main.go:732 fs_type=EXT4_SUPER_MAGIC
prometheus_1       | level=info ts=2020-10-29T07:15:09.210Z caller=main.go:735 msg="TSDB started"
prometheus_1       | level=info ts=2020-10-29T07:15:09.210Z caller=main.go:861 msg="Loading configuration file" filename=/etc/prometheus/prometheus.yml
prometheus_1       | level=info ts=2020-10-29T07:15:09.214Z caller=main.go:892 msg="Completed loading of configuration file" filename=/etc/prometheus/prometheus.yml totalDuration=3.5797ms remote_storage=8.9µs web_handler=7.4µs query_engine=7.5µs scrape=330.7µs scrape_sd=72.1µs notify=6.4µs notify_sd=10µs rules=8.2µs
prometheus_1       | level=info ts=2020-10-29T07:15:09.214Z caller=main.go:684 msg="Server is ready to receive web requests."
{{< / highlight >}}

That's all, the exporter should be up and running. Have fun with it and hopefully you will gather interesting metrics and never run into issues. You can access the exporter at [http://localhost:9504/metrics](http://localhost:9504/metrics) and [Prometheus](https://prometheus.io) at [http://localhost:9090](http://localhost:9090).

## Configuration

GITHUB_EXPORTER_LOG_LEVEL
: Only log messages with given severity, defaults to `info`

GITHUB_EXPORTER_LOG_PRETTY
: Enable pretty messages for logging, defaults to `false`

GITHUB_EXPORTER_WEB_ADDRESS
: Address to bind the metrics server, defaults to `0.0.0.0:9504`

GITHUB_EXPORTER_WEB_PATH
: Path to bind the metrics server, defaults to `/metrics`

GITHUB_EXPORTER_TLS_INSECURE
: Skip host verify on TLS connection, defaults to `false`

GITHUB_EXPORTER_REQUEST_TIMEOUT
: Request to Github timeout, defaults to `5s`

GITHUB_EXPORTER_SERVER_TIMEOUT
: Prometheus /metrics endpoint timeout, defaults to `10s`

GITHUB_EXPORTER_TOKEN
: Access token for the GitHub API

GITHUB_EXPORTER_BASE_URL
: URL to access the GitHub API, defaults to `https://api.github.com/`

GITHUB_EXPORTER_ENTERPRISE
: Enterprises to scrape metrics from, comma-separated list

GITHUB_EXPORTER_ORG
: Organizations to scrape metrics from, comma-separated list

GITHUB_EXPORTER_REPO
: Repositories to scrape metrics from, comma-separated list

GITHUB_EXPORTER_COLLECTOR_ORGS
: Enable collector for orgs, defaults to  `true`

GITHUB_EXPORTER_COLLECTOR_REPOS
: Enable collector for repos, defaults to `true`

GITHUB_EXPORTER_COLLECTOR_ACTIONS
: Enable collector for actions, defaults to  `false`

GITHUB_EXPORTER_COLLECTOR_PACKAGES
: Enable collector for packages, defaults to  `false`

GITHUB_EXPORTER_COLLECTOR_STORAGE
: Enable collector for storage, defaults to  `false`

## Metrics

{{< highlight txt >}}
# HELP github_action_billing_included_minutes Included minutes for this type
# TYPE github_action_billing_included_minutes gauge
github_action_billing_included_minutes{name="promhippie",type="org"} 2000
# HELP github_action_billing_minutes_used Total action minutes used for this type
# TYPE github_action_billing_minutes_used gauge
github_action_billing_minutes_used{name="promhippie",type="org"} 0
# HELP github_action_billing_paid_minutes Total paid minutes used for this type
# TYPE github_action_billing_paid_minutes gauge
github_action_billing_paid_minutes{name="promhippie",type="org"} 0
# HELP github_build_info A metric with a constant '1' value labeled by version, revision and goversion from which it was built.
# TYPE github_build_info gauge
github_build_info{goversion="go1.15.3",revision="89643fa",version="89643fa"} 1
# HELP github_org_collaborators Number of collaborators within org
# TYPE github_org_collaborators gauge
github_org_collaborators{name="promhippie"} 0
# HELP github_org_create_timestamp Timestamp of the creation of org
# TYPE github_org_create_timestamp gauge
github_org_create_timestamp{name="promhippie"} 1.536827847e+09
# HELP github_org_disk_usage Used diskspace by the org
# TYPE github_org_disk_usage gauge
github_org_disk_usage{name="promhippie"} 7392
# HELP github_org_followers Number of followers for org
# TYPE github_org_followers gauge
github_org_followers{name="promhippie"} 0
# HELP github_org_following Number of following other users by org
# TYPE github_org_following gauge
github_org_following{name="promhippie"} 0
# HELP github_org_private_gists Number of private gists from org
# TYPE github_org_private_gists gauge
github_org_private_gists{name="promhippie"} 0
# HELP github_org_private_repos_owned Owned private repositories by org
# TYPE github_org_private_repos_owned gauge
github_org_private_repos_owned{name="promhippie"} 0
# HELP github_org_private_repos_total Total amount of private repositories
# TYPE github_org_private_repos_total gauge
github_org_private_repos_total{name="promhippie"} 0
# HELP github_org_public_gists Number of public gists from org
# TYPE github_org_public_gists gauge
github_org_public_gists{name="promhippie"} 0
# HELP github_org_public_repos Number of public repositories from org
# TYPE github_org_public_repos gauge
github_org_public_repos{name="promhippie"} 11
# HELP github_org_updated_timestamp Timestamp of the last modification of org
# TYPE github_org_updated_timestamp gauge
github_org_updated_timestamp{name="promhippie"} 1.603206498e+09
# HELP github_package_billing_gigabytes_bandwidth_used Total bandwidth used by this type in Gigabytes
# TYPE github_package_billing_gigabytes_bandwidth_used gauge
github_package_billing_gigabytes_bandwidth_used{name="promhippie",type="org"} 0
# HELP github_package_billing_included_gigabytes_bandwidth Included bandwidth for this type in Gigabytes
# TYPE github_package_billing_included_gigabytes_bandwidth gauge
github_package_billing_included_gigabytes_bandwidth{name="promhippie",type="org"} 1
# HELP github_package_billing_paid_gigabytes_bandwidth_used Total paid bandwidth used by this type in Gigabytes
# TYPE github_package_billing_paid_gigabytes_bandwidth_used gauge
github_package_billing_paid_gigabytes_bandwidth_used{name="promhippie",type="org"} 0
# HELP github_repo_allow_merge_commit Show if this repository allows merge commits
# TYPE github_repo_allow_merge_commit gauge
github_repo_allow_merge_commit{name="alpine",owner="dockhippie"} 1
github_repo_allow_merge_commit{name="debian",owner="dockhippie"} 1
github_repo_allow_merge_commit{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_allow_rebase_merge Show if this repository allows rebase merges
# TYPE github_repo_allow_rebase_merge gauge
github_repo_allow_rebase_merge{name="alpine",owner="dockhippie"} 1
github_repo_allow_rebase_merge{name="debian",owner="dockhippie"} 1
github_repo_allow_rebase_merge{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_allow_squash_merge Show if this repository allows squash merges
# TYPE github_repo_allow_squash_merge gauge
github_repo_allow_squash_merge{name="alpine",owner="dockhippie"} 1
github_repo_allow_squash_merge{name="debian",owner="dockhippie"} 1
github_repo_allow_squash_merge{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_archived Show if this repository have been archived
# TYPE github_repo_archived gauge
github_repo_archived{name="alpine",owner="dockhippie"} 0
github_repo_archived{name="debian",owner="dockhippie"} 0
github_repo_archived{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_created_timestamp Timestamp of the creation of repo
# TYPE github_repo_created_timestamp gauge
github_repo_created_timestamp{name="alpine",owner="dockhippie"} 1.424781107e+09
github_repo_created_timestamp{name="debian",owner="dockhippie"} 1.431435933e+09
github_repo_created_timestamp{name="ubuntu",owner="dockhippie"} 1.451171276e+09
# HELP github_repo_forked Show if this repository is a forked repository
# TYPE github_repo_forked gauge
github_repo_forked{name="alpine",owner="dockhippie"} 0
github_repo_forked{name="debian",owner="dockhippie"} 0
github_repo_forked{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_forks How often has this repository been forked
# TYPE github_repo_forks gauge
github_repo_forks{name="alpine",owner="dockhippie"} 13
github_repo_forks{name="debian",owner="dockhippie"} 0
github_repo_forks{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_has_downloads Show if this repository got downloads enabled
# TYPE github_repo_has_downloads gauge
github_repo_has_downloads{name="alpine",owner="dockhippie"} 0
github_repo_has_downloads{name="debian",owner="dockhippie"} 0
github_repo_has_downloads{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_has_issues Show if this repository got issues enabled
# TYPE github_repo_has_issues gauge
github_repo_has_issues{name="alpine",owner="dockhippie"} 1
github_repo_has_issues{name="debian",owner="dockhippie"} 1
github_repo_has_issues{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_has_pages Show if this repository got pages enabled
# TYPE github_repo_has_pages gauge
github_repo_has_pages{name="alpine",owner="dockhippie"} 0
github_repo_has_pages{name="debian",owner="dockhippie"} 0
github_repo_has_pages{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_has_projects Show if this repository got projects enabled
# TYPE github_repo_has_projects gauge
github_repo_has_projects{name="alpine",owner="dockhippie"} 1
github_repo_has_projects{name="debian",owner="dockhippie"} 1
github_repo_has_projects{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_has_wiki Show if this repository got wiki enabled
# TYPE github_repo_has_wiki gauge
github_repo_has_wiki{name="alpine",owner="dockhippie"} 0
github_repo_has_wiki{name="debian",owner="dockhippie"} 0
github_repo_has_wiki{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_issues Number of open issues on this repository
# TYPE github_repo_issues gauge
github_repo_issues{name="alpine",owner="dockhippie"} 1
github_repo_issues{name="debian",owner="dockhippie"} 0
github_repo_issues{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_network Number of repositories in the network
# TYPE github_repo_network gauge
github_repo_network{name="alpine",owner="dockhippie"} 13
github_repo_network{name="debian",owner="dockhippie"} 0
github_repo_network{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_private Show iof this repository is private
# TYPE github_repo_private gauge
github_repo_private{name="alpine",owner="dockhippie"} 0
github_repo_private{name="debian",owner="dockhippie"} 0
github_repo_private{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_pushed_timestamp Timestamp of the last push to repo
# TYPE github_repo_pushed_timestamp gauge
github_repo_pushed_timestamp{name="alpine",owner="dockhippie"} 1.601480354e+09
github_repo_pushed_timestamp{name="debian",owner="dockhippie"} 1.60259358e+09
github_repo_pushed_timestamp{name="ubuntu",owner="dockhippie"} 1.603514407e+09
# HELP github_repo_size Size of the repository content
# TYPE github_repo_size gauge
github_repo_size{name="alpine",owner="dockhippie"} 27393
github_repo_size{name="debian",owner="dockhippie"} 30119
github_repo_size{name="ubuntu",owner="dockhippie"} 30174
# HELP github_repo_stargazers Number of stargazers on this repository
# TYPE github_repo_stargazers gauge
github_repo_stargazers{name="alpine",owner="dockhippie"} 9
github_repo_stargazers{name="debian",owner="dockhippie"} 0
github_repo_stargazers{name="ubuntu",owner="dockhippie"} 0
# HELP github_repo_subscribers Number of subscribers on this repository
# TYPE github_repo_subscribers gauge
github_repo_subscribers{name="alpine",owner="dockhippie"} 4
github_repo_subscribers{name="debian",owner="dockhippie"} 1
github_repo_subscribers{name="ubuntu",owner="dockhippie"} 1
# HELP github_repo_updated_timestamp Timestamp of the last modification of repo
# TYPE github_repo_updated_timestamp gauge
github_repo_updated_timestamp{name="alpine",owner="dockhippie"} 1.5994306e+09
github_repo_updated_timestamp{name="debian",owner="dockhippie"} 1.602593578e+09
github_repo_updated_timestamp{name="ubuntu",owner="dockhippie"} 1.603514408e+09
# HELP github_repo_watchers Number of watchers on this repository
# TYPE github_repo_watchers gauge
github_repo_watchers{name="alpine",owner="dockhippie"} 9
github_repo_watchers{name="debian",owner="dockhippie"} 0
github_repo_watchers{name="ubuntu",owner="dockhippie"} 0
# HELP github_request_failures_total Total number of failed requests to the GitHub API per collector.
# TYPE github_request_failures_total counter
github_request_failures_total{collector="action"} 0
github_request_failures_total{collector="org"} 0
github_request_failures_total{collector="package"} 0
github_request_failures_total{collector="repo"} 0
github_request_failures_total{collector="storage"} 0
# HELP github_storage_billing_days_left_in_cycle Days left within this billing cycle for this type
# TYPE github_storage_billing_days_left_in_cycle gauge
github_storage_billing_days_left_in_cycle{name="promhippie",type="org"} 22
# HELP github_storage_billing_estimated_paid_storage_for_month Estimated paid storage for this month for this type
# TYPE github_storage_billing_estimated_paid_storage_for_month gauge
github_storage_billing_estimated_paid_storage_for_month{name="promhippie",type="org"} 0
# HELP github_storage_billing_estimated_storage_for_month Estimated total storage for this month for this type
# TYPE github_storage_billing_estimated_storage_for_month gauge
github_storage_billing_estimated_storage_for_month{name="promhippie",type="org"} 0
{{< / highlight >}}
