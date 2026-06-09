FROM --platform=$BUILDPLATFORM golang:1.26.4-alpine@sha256:f23e8b227fb4493eabe03bede4d5a32d04092da71962f1fb79b5f7d1e6c2a17f AS builder

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

FROM alpine:3.24@sha256:660e0827bd401543d81323d4886abbd08fda0fe3ba84337837d0b11a67251283

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
