# GitHub Exporter

[![Current Tag](https://img.shields.io/github/v/tag/promhippie/github_exporter?sort=semver)](https://github.com/promhippie/github_exporter) [![General Build](https://github.com/promhippie/github_exporter/actions/workflows/general.yml/badge.svg)](https://github.com/promhippie/github_exporter/actions/workflows/general.yml) [![Join the Matrix chat at https://matrix.to/#/#webhippie:matrix.org](https://img.shields.io/badge/matrix-%23webhippie-7bc9a4.svg)](https://matrix.to/#/#webhippie:matrix.org) [![Codacy Badge](https://app.codacy.com/project/badge/Grade/af9b80ac46294ac9a52d823e991eb4e9)](https://app.codacy.com/gh/promhippie/github_exporter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Go Doc](https://godoc.org/github.com/promhippie/github_exporter?status.svg)](http://godoc.org/github.com/promhippie/github_exporter) [![Go Report](http://goreportcard.com/badge/github.com/promhippie/github_exporter)](http://goreportcard.com/report/github.com/promhippie/github_exporter) [![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com)

An exporter for [Prometheus][prometheus] that collects metrics from
[GitHub][github].

## Install

You can download prebuilt binaries from our [GitHub releases][releases]. Besides
that we also prepared repositories for DEB and RPM packages which can be found
at [Cloudsmith][pkgrepo]. If you prefer to use containers you could use our
images published on [GHCR][ghcr], [Docker Hub][dockerhub] or [Quay][quayio]. If
you need further guidance how to install this take a look at our [docs][docs].

Package repository hosting is graciously provided by [Cloudsmith][cloudsmith].
Cloudsmith is the only fully hosted, cloud-native, universal package management
solution, that enables your organization to create, store and share packages in
any format, to any place, with total confidence.

## Development

If you are not familiar with [Nix][nix] it is up to you to have a working
environment for Go (>= 1.24.0) as the setup won't we covered within this guide.
Please follow the official install instructions for [Go][golang]. Beside that
we are using [go-task][gotask] to define all commands to build this project.

```console
git clone https://github.com/promhippie/github_exporter.git
cd github_exporter

task generate build
./bin/github_exporter -h
```

If you got [Nix][nix] and [Direnv][direnv] configured you can simply execute
the following commands to get al dependencies including [go-task][gotask] and
the required runtimes installed. You are also able to directly use the process
manager of [devenv][devenv]:

```console
cat << EOF > .envrc
use flake . --impure --extra-experimental-features nix-command
EOF

direnv allow
```

To start developing on this project you have to execute only a few commands:

```console
task watch
```

The development server should be running on
[http://localhost:9504](http://localhost:9504). Generally it supports
hot reloading which means the services are automatically restarted/reloaded on
code changes.

If you got [Nix][nix] configured you can simply execute the [devenv][devenv]
command to start:

```console
devenv up
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
[pkgrepo]: https://cloudsmith.io/~webhippie/repos/promhippie/groups/
[cloudsmith]: https://cloudsmith.com/
[ghcr]: https://github.com/promhippie/github_exporter/pkgs/container/github_exporter
[dockerhub]: https://hub.docker.com/r/promhippie/github-exporter/tags/
[quayio]: https://quay.io/repository/promhippie/github-exporter?tab=tags
[docs]: https://promhippie.github.io/github_exporter/#getting-started
[nix]: https://nixos.org/
[golang]: http://golang.org/doc/install.html
[gotask]: https://taskfile.dev/installation/
[direnv]: https://direnv.net/
[devenv]: https://devenv.sh/
