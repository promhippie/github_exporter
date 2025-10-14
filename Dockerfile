FROM --platform=$BUILDPLATFORM golang:1.25.3-alpine3.21@sha256:2c9684db68f1b6e76a500fdb1ea9af6288725b7f3ef47aa3265195d3ed5a8326 AS builder

RUN apk add --no-cache -U git curl
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /go/src/exporter
COPY . /go/src/exporter/

RUN --mount=type=cache,target=/go/pkg \
    go mod download -x

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    task generate build GOOS=${TARGETOS} GOARCH=${TARGETARCH}

FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

RUN apk add --no-cache ca-certificates mailcap && \
    addgroup -g 1337 exporter && \
    adduser -D -u 1337 -h /var/lib/exporter -G exporter exporter

EXPOSE 9504
VOLUME ["/var/lib/exporter"]
ENTRYPOINT ["/usr/bin/github_exporter"]
HEALTHCHECK CMD ["/usr/bin/github_exporter", "health"]

ENV GITHUB_EXPORTER_DATABASE_DSN=sqlite:///var/lib/exporter/database.sqlite3

COPY --from=builder /go/src/exporter/bin/github_exporter /usr/bin/github_exporter
WORKDIR /var/lib/exporter
USER exporter
