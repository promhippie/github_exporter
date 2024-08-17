# Compose

With the `docker-compose` snippets within this directory you are able to plug
different setups of Gopad together. Below you can find some example
combinations.

## Base

First of all we need the base definition and we need to decide if we want to
build the Docker image dynamically or if we just want to use a released Docker
image.

### Build

This simply takes the currently cloned source and builds a new Docker image
including all local changes.

```console
docker-compose -f hack/compose/base.yml -f hack/compose/build.yml up
```

### Image

This simply downloads the defined image from DockerHub to  start and configure
it properly.

```console
docker-compose -f hack/compose/base.yml -f hack/compose/image.yml up
```

## Parca

To gather some insights about the memory allocation and the cpu usage you could
optionally launcher [Parca][parca] to continuously fetch pprof details. You can
access [Parca][parca] on [http://localhost:7070](http://localhost:7070).

```console
docker-compose <base from above> -f hack/compose/parca.yml up
```

## Metrics

To launch a stack of [Prometheus][prometheus] and [Grafana][grafana] to directly
visualizing the metrics you can just add this to the command. You can access
[Prometheus][prometheus] on [http://localhost:9090](http://localhost:9090) and
[Grafana][grafana] on [http://localhost:3000](http://localhost:3000).

```console
docker-compose <base from above> -f hack/compose/metrics.yml up
```

## Database

Finally you can enable the database support for this exporter which is required
to work with workflow events sent from GitHub via webhook. Here we got currently
the following options so far.

### SQLite

This simply configures a named volume for the SQLite storage used as a database
backend.

```console
docker-compose <base from above> -f hack/compose/db/sqlite.yml up
```

### MariaDB

This simply starts an additional container for a MariaDB instance used as a
database backend.

```console
docker-compose <base from above> -f hack/compose/db/mariadb.yml up
```

### PostgreSQL

This simply starts an additional container for a PostgreSQL instance used as a
database backend.

```console
docker-compose <base from above> -f hack/compose/db/postgres.yml up
```

[parca]: https://www.parca.dev/
[prometheus]: https://prometheus.io/
[grafana]: https://grafana.com/
