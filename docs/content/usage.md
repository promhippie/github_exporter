---
title: "Usage"
date: 2023-10-20T00:00:00+00:00
anchor: "getting-started"
weight: 10
---

## Installation

We won't cover further details how to properly setup [Prometheus][prometheus]
itself, we will only cover some basic setup based on [docker-compose][compose].
But if you want to run this exporter without [docker-compose][compose] you
should be able to adopt that to your needs.

First of all we need to prepare a configuration for [Prometheus][prometheus]
that includes the exporter based on a static configuration with the container
name as a hostname:

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

After preparing the configuration we need to create the `docker-compose.yml`
within the same folder, this `docker-compose.yml` starts a simple
[Prometheus][prometheus] instance together with the exporter. Don't forget to
update the environment variables with the required credentials.

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
    image: promhippie/github-exporter:latest
    restart: always
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

Since our `latest` tag always refers to the `master` branch of the Git
repository you should always use some fixed version. You can see all available
tags at [DockerHub][dockerhub] or [Quay][quayio], there you will see that we
also provide a manifest, you can easily start the exporter on various
architectures without any change to the image name. You should apply a change
like this to the `docker-compose.yml` file:

{{< highlight diff >}}
  github_exporter:
-   image: promhippie/github-exporter:latest
+   image: promhippie/github-exporter:1.0.0
    restart: always
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

If you want to access the exporter directly you should bind it to a local port,
otherwise only [Prometheus][prometheus] will have access to the exporter. For
debugging purpose or just to discover all available metrics directly you can
apply this change to your `docker-compose.yml`, after that you can access it
directly at [http://localhost:9504/metrics](http://localhost:9504/metrics):

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
+   ports:
+     - 127.0.0.1:9504:9504
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

It's also possible to provide the token to access the GitHub API gets provided
by a file, in case you are using some kind of secret provider. For this use case
you can write the token to a file on any path and reference it with the
following format:

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
-     - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
+     - GITHUB_EXPORTER_TOKEN=file://path/to/secret/file/with/token
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

Besides the `file://` format we currently also support `base64://` which expects
the token in a base64 encoded format. This functionality can be used for the
token and other secret values like the private key for GitHub App authentication
so far.

If you want to collect the metrics of all repositories within an organization
you are able to use globbing, but be aware that all repositories matched by
globbing won't provide metrics for the number of subscribers, the number of
repositories in the network, if squash merges are allowed, if rebase merges are
allowed or merge commits are allowed. These metrics are only present for
specific repositories like the example mentioned above.

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
-     - GITHUB_EXPORTER_REPO=promhippie/example
+     - GITHUB_EXPORTER_REPO=promhippie/*_exporter,promhippie/prometheus*
{{< / highlight >}}

If you want to secure the access to the exporter you can provide a web config.
You just need to provide a path to the config file in order to enable the
support for it, for details about the config format look at the
[documentation](#web-configuration) section:

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
+     - GITHUB_EXPORTER_WEB_CONFIG=path/to/web-config.json
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

If you want to use the workflows exporter you are forced to expose the exporter
on the internet as the exporter got to receive webhooks from GitHub. Otherwise
you won't be able to receive information about the workflows which could be
transformed to metrics.

To enable the webhook endpoint you should prepare a random secret which gets
used by the endpoint and the GitHub webhook, it can have any format and any
length.

Make sure that the exporter is reachable on the `/github` endpoint and add the
following environment variables, best would be to use some kind of reverse proxy
in front of the exporter which enforces connections via HTTPS or to properly
configure HTTPS access via [web configuration](#web-configuration).

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
+     - GITHUB_EXPORTER_COLLECTOR_WORKFLOWS=true
+     - GITHUB_EXPORTER_WEBHOOK_SECRET=your-prepared-random-secret
      - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

After you have enabled the workflow collector and made sure that the endpoint is
reachable by GitHub you can look at the [webhook](#webhook) section of this
documentation to see how to configure the webhook on your GitHub organization
or repository.

If you want to use a GitHub application instead of a personal access token
please take a look at the [application](#application) section and add the
following environment variables after that:

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
+     - GITHUB_EXPORTER_APP_ID=your-application-id
+     - GITHUB_EXPORTER_INSTALLATION_ID=your-installation-id
+     - GITHUB_EXPORTER_PRIVATE_KEY=file://path/to/secret.pem
-     - GITHUB_EXPORTER_TOKEN=bldyecdtysdahs76ygtbw51w3oeo6a4cvjwoitmb
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

If you prefer to provide the private key as a string instead you could also
provide is in a base64 encoded format:

{{< highlight diff >}}
  github_exporter:
    image: promhippie/github-exporter:latest
    restart: always
    environment:
      - GITHUB_EXPORTER_APP_ID=your-application-id
      - GITHUB_EXPORTER_INSTALLATION_ID=your-installation-id
+     - GITHUB_EXPORTER_PRIVATE_KEY=base64://Q0VSVElGSUNBVEU=
-     - GITHUB_EXPORTER_PRIVATE_KEY=file://path/to/secret.pem
      - GITHUB_EXPORTER_LOG_PRETTY=true
      - GITHUB_EXPORTER_ORG=promhippie
      - GITHUB_EXPORTER_REPO=promhippie/example
{{< / highlight >}}

Finally the exporter should be configured fine, let's start this stack with
[docker-compose][compose], you just need to execute `docker-compose up` within
the directory where you have stored the `prometheus.yml` and
`docker-compose.yml`.

That's all, the exporter should be up and running. Have fun with it and
hopefully you will gather interesting metrics and never run into issues. You can
access the exporter at
[http://localhost:9504/metrics](http://localhost:9504/metrics) and
[Prometheus][prometheus] at [http://localhost:9090](http://localhost:9090).

## Configuration

{{< partial "envvars.md" >}}

### Web Configuration

If you want to secure the service by TLS or by some basic authentication you can
provide a `YAML` configuration file which follows the [Prometheus][prometheus]
toolkit format. You can see a full configuration example within the
[toolkit documentation][toolkit].

## Metrics

You can a rough list of available metrics below, additionally to these metrics
you will always get the standard metrics exported by the Golang client of
[Prometheus][prometheus]. If you want to know more about these standard metrics
take a look at the [process collector][proccollector] and the
[Go collector][gocollector].

{{< partial "metrics.md" >}}

[prometheus]: https://prometheus.io
[compose]: https://docs.docker.com/compose/
[dockerhub]: https://hub.docker.com/r/promhippie/github-exporter/tags/
[quayio]: https://quay.io/repository/promhippie/github-exporter?tab=tags
[toolkit]: https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
[proccollector]: https://github.com/prometheus/client_golang/blob/master/prometheus/process_collector.go
[gocollector]: https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go
