---
title: "Getting Started"
date: 2020-10-28T00:00:00+00:00
anchor: "getting-started"
weight: 10
---

## Installation

We won't cover further details how to properly setup [Prometheus](https://prometheus.io) itself, we will only cover some basic setup based on [docker-compose](https://docs.docker.com/compose/). But if you want to run this exporter without [docker-compose](https://docs.docker.com/compose/) you should be able to adopt that to your needs.

First of all we need to prepare a configuration for [Prometheus](https://prometheus.io) that includes the exporter as a target based on a static host mapping which is just the [docker-compose](https://docs.docker.com/compose/) container name, e.g. `github-exporter`.

{{< highlight yaml >}}
global:
  scrape_interval: 1m
  scrape_timeout: 10s
  evaluation_interval: 1m

scrape_configs:
- job_name: github
  static_configs:
  - targets:
    - github-exporter:9504
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

  github-exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

Since our `latest` Docker tag always refers to the `master` branch of the Git repository you should always use some fixed version. You can see all available tags at our [DockerHub repository](https://hub.docker.com/r/promhippie/github-exporter/tags/), there you will see that we also provide a manifest, you can easily start the exporter on various architectures without any change to the image name. You should apply a change like this to the `docker-compose.yml`:

{{< highlight diff >}}
  hcloud-exporter:
-   image: promhippie/github-exporter:latest
+   image: promhippie/github-exporter:0.1.0
    restart: always
    environment:
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

If you want to access the exporter directly you should bind it to a local port, otherwise only [Prometheus](https://prometheus.io) will have access to the exporter. For debugging purpose or just to discover all available metrics directly you can apply this change to your `docker-compose.yml`, after that you can access it directly at [http://localhost:9504/metrics](http://localhost:9504/metrics):

{{< highlight diff >}}
  hcloud-exporter:
    image: promhippie/github-exporter:latest
    restart: always
+   ports:
+     - 127.0.0.1:9504:9504
    environment:
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
{{< / highlight >}}

Finally the exporter should be configured fine, let's start this stack with [docker-compose](https://docs.docker.com/compose/), you just need to execute `docker-compose up` within the directory where you have stored the `prometheus.yml` and `docker-compose.yml`.

{{< highlight txt >}}
TBD
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

GITHUB_EXPORTER_REQUEST_TIMEOUT
: Request timeout as duration, defaults to `5s`

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

TBD
