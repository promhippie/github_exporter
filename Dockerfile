FROM --platform=$BUILDPLATFORM golang:1.25.0-alpine3.21@sha256:9030d461ce2f9586ad6fac3c17607b3fb3df38c7d9f6e567aa05ea716fb107d8 AS builder

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

FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

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
