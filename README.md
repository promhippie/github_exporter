# GitHub Exporter

[![Build Status](http://github.dronehippie.de/api/badges/webhippie/github_exporter/status.svg)](http://github.dronehippie.de/webhippie/github_exporter)
[![Go Doc](https://godoc.org/github.com/webhippie/github_exporter?status.svg)](http://godoc.org/github.com/webhippie/github_exporter)
[![Go Report](http://goreportcard.com/badge/github.com/webhippie/github_exporter)](http://goreportcard.com/report/github.com/webhippie/github_exporter)
[![](https://images.microbadger.com/badges/image/tboerger/github-exporter.svg)](http://microbadger.com/images/tboerger/github-exporter "Get your own image badge on microbadger.com")
[![Join the chat at https://gitter.im/webhippie/general](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/webhippie/general)

A [Prometheus](https://prometheus.io/) exporter that collects GitHub statistics for defined namespaces and repositories.


## Installation

If you are missing something just write us on our nice [Gitter](https://gitter.im/webhippie/general) chat. If you find a security issue please contact thomas@webhippie.de first. Currently we are providing only a Docker image at `tboerger/github-exporter`.


### Usage

```bash
# docker run -ti --rm tboerger/github-exporter -h
Usage of /bin/github_exporter:
  -github.org value
      Organizations to watch on GitHub
  -github.repo value
      Repositories to watch on GitHub
  -log.format value
      Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true" (default "logger:stderr")
  -log.level value
      Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal] (default "info")
  -version
      Print version information
  -web.listen-address string
      Address to listen on for web interface and telemetry (default ":9104")
  -web.telemetry-path string
      Path to expose metrics of the exporter (default "/metrics")
```


## Development

Make sure you have a working Go environment, for further reference or a guide take a look at the [install instructions](http://golang.org/doc/install.html). It is also possible to just simply execute the `go get github.com/webhippie/github_exporter` command, but we prefer to use our `Makefile`:

```bash
go get -d github.com/webhippie/github_exporter
cd $GOPATH/src/github.com/webhippie/github_exporter
make test build

./github_exporter -h
```


## Metrics

```
# HELP github_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which github_exporter was built.
# TYPE github_exporter_build_info gauge
github_exporter_build_info{branch="HEAD",goversion="go1.8.1",revision="970771d5cbf98d9a3347f0a15e3e9438ebb5cfe4",version="0.2.0"} 1
# HELP github_forks How often have this repository been forked
# TYPE github_forks gauge
github_forks{owner="webhippie",repo="redirects"} 0
# HELP github_issues How many open issues does the repository have
# TYPE github_issues gauge
github_issues{owner="webhippie",repo="redirects"} 5
# HELP github_pushed A timestamp when the repository had the last push
# TYPE github_pushed gauge
github_pushed{owner="webhippie",repo="redirects"} 1.49328094e+09
# HELP github_size Simply the size of the Git repository
# TYPE github_size gauge
github_size{owner="webhippie",repo="redirects"} 1769
# HELP github_stars How often have this repository been stared
# TYPE github_stars gauge
github_stars{owner="webhippie",repo="redirects"} 2
# HELP github_up Check if GitHub response can be processed
# TYPE github_up gauge
github_up 1
# HELP github_updated A timestamp when the repository have been updated
# TYPE github_updated gauge
github_updated{owner="webhippie",repo="redirects"} 1.492589212e+09
# HELP github_watchers How often have this repository been watched
# TYPE github_watchers gauge
github_watchers{owner="webhippie",repo="redirects"} 2
```


## Contributing

Fork -> Patch -> Push -> Pull Request


## Authors

* [Thomas Boerger](https://github.com/tboerger)


## License

Apache-2.0


## Copyright

```
Copyright (c) 2017 Thomas Boerger <http://www.webhippie.de>
```
