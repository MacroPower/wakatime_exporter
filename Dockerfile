ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Jacob Colvin (MacroPower) <me@jacobcolvin.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/wakatime_exporter /bin/wakatime_exporter

USER nobody
ENTRYPOINT ["/bin/wakatime_exporter"]
EXPOSE 9212
