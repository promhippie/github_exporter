FROM quay.io/prometheus/busybox:latest
MAINTAINER Thomas Boerger <thomas@webhippie.de>

COPY github_exporter /bin/github_exporter

EXPOSE 9104
ENTRYPOINT ["/bin/github_exporter"]
