# Alces Concertim Metric Reporting and Processing Daemons

## Overview

This repository contains two daemons.

Concertim Metric Reporting Daemon (ct-metric-reporting-daemon) provides a HTTP
API for metrics to be reported to Concertim and exposes reported metrics to
Ganglia's Gmetad via TCP.

Concertim Metric Processing Daemon (ct-metric-processing-daemon) periodically
polls Ganglia's Gmetad server for metrics, creates useful views of those
metrics and places the most recent views in memcache.

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
