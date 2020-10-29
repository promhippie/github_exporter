# GitHub Exporter

[![Build Status](http://cloud.drone.io/api/badges/promhippie/github_exporter/status.svg)](http://cloud.drone.io/promhippie/github_exporter)
[![Join the Matrix chat at https://matrix.to/#/#webhippie:matrix.org](https://img.shields.io/badge/matrix-%23webhippie-7bc9a4.svg)](https://matrix.to/#/#webhippie:matrix.org)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/af9b80ac46294ac9a52d823e991eb4e9)](https://www.codacy.com/gh/promhippie/github_exporter/dashboard?utm_source=github.com&utm_medium=referral&utm_content=promhippie/github_exporter&utm_campaign=Badge_Grade)
[![Go Doc](https://godoc.org/github.com/promhippie/github_exporter?status.svg)](http://godoc.org/github.com/promhippie/github_exporter)
[![Go Report](http://goreportcard.com/badge/github.com/promhippie/github_exporter)](http://goreportcard.com/report/github.com/promhippie/github_exporter)
[![](https://images.microbadger.com/badges/image/promhippie/github_exporter.svg)](http://microbadger.com/images/promhippie/github_exporter "Get your own image badge on microbadger.com")

An exporter for [Prometheus](https://prometheus.io/) that collects metrics from [GitHub](https://github.com).

## Install

You can download prebuilt binaries from our [GitHub releases](https://github.com/promhippie/github_exporter/releases), or you can use our Docker images published on [Docker Hub](https://hub.docker.com/r/promhippie/github_exporter/tags/). If you need further guidance how to install this take a look at our [documentation](https://promhippie.github.io/github_exporter/#getting-started).

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
