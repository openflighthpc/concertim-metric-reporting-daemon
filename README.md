# Alces Concertim Metric Reporting and Processing Daemon

## Overview

Concertim Metric Reporting Daemon (ct-metric-reporting-daemon) provides a HTTP
API for metrics to be reported to Concertim.

It exposes reported metrics to Ganglia's Gmetad via TCP, it periodically polls
Ganglia's Gmetad server for metrics and creates "views" of those metrics. The
metrics and the associated views are made available over a HTTP API.

Currently the "views" of the metrics include a list of unique metrics and for
each metric a list of values for each device currently reporting that metric.

The use of Ganglia's Gmetad creates RRD files and expires old metrics.

## Usage

See the [usage docs](docs/usage.md) for details on using Concertim Metric
Reporting Daemon.

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
