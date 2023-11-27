# syntax=docker/dockerfile:1.4

###################################
# Build the application from source
FROM golang:1.21 AS build

WORKDIR /app
COPY . /app
RUN mkdir -p /var/lib/metric-reporting-daemon/rrds
RUN make clean
RUN make ct-metric-reporting-daemon

###################################
# Dev stage
FROM build AS dev

ENV GO_ENUM_VERSION=v0.5.8
RUN go install github.com/cosmtrek/air@latest \
    && go install github.com/jmattheis/goverter/cmd/goverter@v0.18.0 \
    && curl -fsSL "https://github.com/abice/go-enum/releases/download/${GO_ENUM_VERSION}/go-enum_$(uname -s)_$(uname -m)" -o $(go env GOPATH)/bin/go-enum \
    && chmod +x $(go env GOPATH)/bin/go-enum

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
         rrdtool \
         gmetad \
    && apt-get clean \
    && rm -rf /usr/share/doc /usr/share/man /var/lib/apt/lists/*

ENTRYPOINT ["/app/docker/entrypoint.dev.sh"]

EXPOSE 3000

###################################
# Run the tests in the container
FROM build AS tests

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
	    rrdtool \
    && apt-get clean \
    && rm -rf /usr/share/doc /usr/share/man /var/lib/apt/lists/*

COPY --from=build /app/testdata/* /app/testdata/
COPY --from=build /app/api/testdata/* /app/api/testdata/
COPY --from=build /app/rrd/testdata/* /app/rrd/testdata/
COPY --from=build /go/pkg /go/pkg/

ARG CACHEBUST=1
RUN go test -v -count=1 -race -shuffle=on ./...

###################
FROM ubuntu:22.04

ARG BUILD_DATE
ARG BUILD_VERSION
ARG BUILD_REVISION

LABEL org.opencontainers.image.created=$BUILD_DATE
LABEL org.opencontainers.image.version=$BUILD_VERSION
LABEL org.opencontainers.image.revision=$BUILD_REVISION
LABEL org.opencontainers.image.title="Alces Concertim Metric Reporting Daemon"

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
         rrdtool \
         gmetad \
    && apt-get clean \
    && rm -rf /usr/share/doc /usr/share/man /var/lib/apt/lists/*

WORKDIR /app
RUN mkdir -p /var/lib/metric-reporting-daemon/rrds

# Add files containing canned responses.
# COPY --from=build /app/testdata/* /app/testdata/

COPY --from=build /app/ct-metric-reporting-daemon /app/ct-metric-reporting-daemon
COPY --from=build /app/config/*.yml /app/config/
COPY --from=build /app/docker/gmetad.conf /etc/ganglia/gmetad.conf
COPY --from=build /app/docker/entrypoint.sh /app/docker/entrypoint.sh
ENTRYPOINT ["/app/docker/entrypoint.sh"]

EXPOSE 3000
