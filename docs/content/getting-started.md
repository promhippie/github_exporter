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

That's all, the exporter should be up and running. Have fun with it and hopefully you will gather interesting metrics and never run into issues. You can access the exporter at [http://localhost:9504/metrics](http://localhost:9504/metrics) and [Prometheus](https://prometheus.io) at [http://localhost:9090](http://localhost:9090).

## Configuration

{{< partial "envvars.md" >}}

## Metrics

You can a rough list of available metrics below, additionally to these metrics you will always get the standard metrics exported by the Golang client of [Prometheus](https://prometheus.io). If you want to know more about these standard metrics take a look at the [process collector](https://github.com/prometheus/client_golang/blob/master/prometheus/process_collector.go) and the [Go collector](https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go).

{{< partial "metrics.md" >}}
