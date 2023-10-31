# syntax=docker/dockerfile:1.4

###################################
# Build the application from source
FROM golang:1.18 AS build-stage
LABEL com.alces-flight.concertim.role=builder

WORKDIR /app
COPY . /app
RUN make clean
RUN make ct-metric-reporting-daemon

###################################
# Run the tests in the container
FROM build-stage AS run-tests
LABEL com.alces-flight.concertim.role=build-test

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
	&& apt-get install --yes --no-install-recommends \
	    rrdtool \
    && apt-get clean \
    && rm -rf /usr/share/doc /usr/share/man /var/lib/apt/lists/*

COPY --from=build-stage /app/testdata/* /app/testdata/
COPY --from=build-stage /app/api/testdata/* /app/api/testdata/
COPY --from=build-stage /app/gds/testdata/* /app/gds/testdata/
COPY --from=build-stage /app/rrd/testdata/* /app/rrd/testdata/
COPY --from=build-stage /go/pkg /go/pkg/

ARG CACHEBUST=1
# Run all tests apart from ticker tests.
RUN go test -v -count=1 -race -shuffle=on ./api ./canned ./config ./domain ./dsmRepository ./gds ./inmem ./processing ./repository/memory ./retrieval ./rrd ./visualizer

###################
FROM ubuntu:22.04
LABEL com.alces-flight.concertim.role=metrics com.alces-flight.concertim.version=0.6.0-dev

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
         rrdtool \
         gmetad \
    && apt-get clean \
    && rm -rf /usr/share/doc /usr/share/man /var/lib/apt/lists/*

WORKDIR /app

# Add files containing canned responses.
# COPY --from=build-stage /app/testdata/* /app/testdata/

COPY --from=build-stage /app/ct-metric-reporting-daemon /app/ct-metric-reporting-daemon
COPY --from=build-stage /app/config/*.yml /app/config/
COPY --from=build-stage /app/docker/gmetad.conf /etc/ganglia/gmetad.conf
COPY --from=build-stage /app/docker/entrypoint.sh /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]

ENV PORT=3000
EXPOSE 3000
