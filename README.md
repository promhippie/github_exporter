# GitHub Exporter

[![Current Tag](https://img.shields.io/github/v/tag/promhippie/github_exporter?sort=semver)](https://github.com/promhippie/prometheus-scw-sd) [![General Build](https://github.com/promhippie/actions/github_exporter/workflows/general.yml/badge.svg)](https://github.com/promhippie/github_exporter/actions/workflows/general.yaml) [![Join the Matrix chat at https://matrix.to/#/#webhippie:matrix.org](https://img.shields.io/badge/matrix-%23webhippie-7bc9a4.svg)](https://matrix.to/#/#webhippie:matrix.org) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/af9b80ac46294ac9a52d823e991eb4e9)](https://www.codacy.com/gh/promhippie/github_exporter/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=promhippie/github_exporter&amp;utm_campaign=Badge_Grade) [![Go Doc](https://godoc.org/github.com/promhippie/github_exporter?status.svg)](http://godoc.org/github.com/promhippie/github_exporter) [![Go Report](http://goreportcard.com/badge/github.com/promhippie/github_exporter)](http://goreportcard.com/report/github.com/promhippie/github_exporter)

An exporter for [Prometheus][prometheus] that collects metrics from
[GitHub][github].

## Install

You can download prebuilt binaries from our [GitHub releases][releases], or you
can use our containers published on [Docker Hub][dockerhub] and [Quay][quayio].
If you need further guidance how to install this take a look at our
[documentation][docs].

## Development

Make sure you have a working Go environment, for further reference or a guide
take a look at the [install instructions][golang]. This project requires
Go >= v1.17, at least that's the version we are using.

```console
git clone https://github.com/promhippie/github_exporter.git
cd github_exporter

make generate build

./bin/github_exporter -h
```

## Security

If you find a security issue please contact
[thomas@webhippie.de](mailto:thomas@webhippie.de) first.

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

[prometheus]: https://prometheus.io
[github]: https://github.com
[releases]: https://github.com/promhippie/github_exporter/releases
[dockerhub]: https://hub.docker.com/r/promhippie/github-exporter/tags/
[quayio]: https://quay.io/repository/promhippie/github-exporter?tab=tags
[docs]: https://promhippie.github.io/github_exporter/#getting-started
[golang]: http://golang.org/doc/install.html
