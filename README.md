# GitHub Exporter

[![Current Tag](https://img.shields.io/github/v/tag/promhippie/github_exporter?sort=semver)](https://github.com/promhippie/github_exporter) [![Build Status](https://drone.webhippie.de/api/badges/promhippie/github_exporter/status.svg)](https://drone.webhippie.de/promhippie/github_exporter) [![Join the Matrix chat at https://matrix.to/#/#webhippie:matrix.org](https://img.shields.io/badge/matrix-%23webhippie-7bc9a4.svg)](https://matrix.to/#/#webhippie:matrix.org) [![Docker Size](https://img.shields.io/docker/image-size/promhippie/github-exporter/latest)](https://hub.docker.com/r/promhippie/github-exporter) [![Docker Pulls](https://img.shields.io/docker/pulls/promhippie/github-exporter)](https://hub.docker.com/r/promhippie/github-exporter) [![Go Reference](https://pkg.go.dev/badge/github.com/promhippie/github_exporter.svg)](https://pkg.go.dev/github.com/promhippie/github_exporter) [![Go Report Card](https://goreportcard.com/badge/github.com/promhippie/github_exporter)](https://goreportcard.com/report/github.com/promhippie/github_exporter) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/af9b80ac46294ac9a52d823e991eb4e9)](https://www.codacy.com/gh/promhippie/github_exporter/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=promhippie/github_exporter&amp;utm_campaign=Badge_Grade)

An exporter for [Prometheus](https://prometheus.io/) that collects metrics from [GitHub](https://github.com).

## Install

You can download prebuilt binaries from our [GitHub releases](https://github.com/promhippie/github_exporter/releases), or you can use our Docker images published on [Docker Hub](https://hub.docker.com/r/promhippie/github-exporter/tags/). If you need further guidance how to install this take a look at our [documentation](https://promhippie.github.io/github_exporter/#getting-started).

## Development

Make sure you have a working Go environment, for further reference or a guide take a look at the [install instructions](http://golang.org/doc/install.html). This project requires Go >= v1.11.

```bash
git clone https://github.com/promhippie/github_exporter.git
cd github_exporter

make generate build

./bin/github_exporter -h
```

## Security

If you find a security issue please contact [thomas@webhippie.de](mailto:thomas@webhippie.de) first.

## Contributing

Fork -> Patch -> Push -> Pull Request

## Authors

-   [Thomas Boerger](https://github.com/tboerger)

## License

Apache-2.0

## Copyright

```console
Copyright (c) 2018 Thomas Boerger <thomas@webhippie.de>
```
