# Alces Concertim Metric Reporting and Processing Daemon

## Overview

Concertim Metric Reporting Daemon (ct-metric-reporting-daemon) provides a HTTP
API for metrics to be reported to Concertim.

The reported metrics are periodically processed by Metric Reporting Daemon.
The processing creates various summaries of the reported metrics and saves them
in [RRD](https://oss.oetiker.ch/rrdtool/index.en.html) files.

The metrics and the associated views are made available over a HTTP API.

## Quick start

1. Clone the repository
    ```bash
    git clone https://github.com/alces-flight/concertim-metric-reporting-daemon.git
    ```
2. Build the docker image
    ```bash
    docker build --network=host --tag concertim-metric-reporting-daemon:latest .
    ```
3. Start the docker container
    ```bash
	docker run -d --name concertim-metric-reporting-daemon \
		--network=host \
		concertim-metric-reporting-daemon
    ```

Use [Concertim OpenStack
Service](https://github.com/alces-flight/concertim-openstack-service) to
collect and report metrics and use [Concertim Visualisation
App](https://github.com/alces-flight/concertim-ct-visualisation-app) to view
the reported metrics.

## Building the docker image

Concertim Metric Reporting Daemon is intended to be deployed as a Docker container.
There is a Dockerfile in this repo for building the image.

1. Clone the repository
    ```bash
    git clone https://github.com/alces-flight/concertim-metric-reporting-daemon.git
    ```
2. Build the docker image
    ```bash
    docker build --network=host --tag concertim-metric-reporting-daemon:latest .
    ```

## Configuration

### Peristent RRD files

The metrics reported to Metric Reporting Daemon are stored in RRD files.  To
ensure that the RRD files outlive the container a volume should be used.  The
example below uses the host directory `/var/lib/metric-reporting-daemon/rrds`.

```bash
sudo mkdir -p /var/lib/metric-reporting-daemon/rrds
```

Mount the directory `/var/lib/metric-reporting-daemon/rrds` to
`/var/lib/metric-reporting-daemon/rrds` when starting the docker container.

```bash
docker run --name concertim-metric-reporting-daemon \
    --network=host \
    --volume /var/lib/metric-reporting-daemon/rrds:/var/lib/metric-reporting-daemon/rrds \
    concertim-metric-reporting-daemon
```

### Configuration file

You can optionally, copy the configuration file to the host, make changes to it
and mount it to the container.

```bash
sudo cp -a config/config.prod.yml /etc/metric-reporting-daemon/config.yml
$EDITOR /etc/metric-reporting-daemon/config.yml
```

Mount the configuration file to
`/opt/concertim/etc/metric-reporting-daemon/config.yml` when starting the
container.

```bash
docker run --name concertim-metric-reporting-daemon \
    --network=host \
    --volume /etc/metric-reporting-daemon/config.yml:/opt/concertim/etc/metric-reporting-daemon/config.yml \
    concertim-metric-reporting-daemon
```

## Usage

Once the docker image has been built, the RRD persistent storage created and
the config file copied to the host, the container can be started with the
following command.

```bash
docker run -d --name concertim-metric-reporting-daemon \
    --network=host \
    --volume /var/lib/metric-reporting-daemon/rrds/:/var/lib/metric-reporting-daemon/rrds/ \
    --volume /etc/metric-reporting-daemon/config.yml:/opt/concertim/etc/metric-reporting-daemon/config.yml \
    concertim-metric-reporting-daemon
```

## HTTP API

See the [usage docs](docs/usage.md) for details on the HTTP API for reporting
and querying metrics.

## Development

See the [development docs](docs/DEVELOPMENT.md) for details on development and
getting started with development.

## Deployment

Concertim Metric Reporting Daemon is deployed as part of the Concertim
appliance using the [Concertim ansible
playbook](https://github.com/alces-flight/concertim-ansible-playbook).


# Copyright and License

GNU Affero General Public License, see [LICENSE.txt](LICENSE.txt) for details.

Copyright (C) 2022-present Stephen F Norledge & Alces Flight Ltd.
