# Development

Concertim Metric Reporting Daemon (variably `ct-metric-reporting-daemon`,
`ct-metrics` or `mrd`) is one of two apps that form the Concertim UI; the other
is
[ct-visualisation-app](https://github.com/alces-flight/concertim-ct-visualisation-app)

## Vagrant machine

Development of both apps takes place on a Vagrant virtual machine, which is
provisioned with the use of an ansible playbook.  The vagrant file and the
ansible playbook can be found in the
[concertim-ansible-playbook](https://github.com/alces-flight/concertim-ansible-playbook)
repo.

That repo contains details on how to build the vagrant machine and provision it
for Concertim development.  Including expectations on where the source code
should be checked out.

## Code structure and architecture

`ct-metrics` is written in Go and provides an HTTP API for metrics to be
reported to Concertim and to retrieve details about those metrics.

Its architecture is influenced by legacy and old-legacy Concertim.  This
influence should be removed in time.

A brief explanation of the directories is as follows:

* `api` package: contains the HTTP API server for reporting and querying
   metrics.  Reported metrics are stored in an implementation of
  `domain.ReportedRepository`.  Metrics are queried from an implementation of
  `domain.ProcessedRepository`.

* `canned` directory contains functionality to provide canned responses for
  querying the data source maps and metrics reported to ganglia's gmetad.

* `cmd` directory contains the executables.

* `config` package contains the config files and config related code.

* `domain` package: contains the domain entities.  Currently an `Application`
  struct, various models to represent both reported and fully processed hosts
  and metrics, and various repository interfaces.

* `dsmRepository` package: contains code for periodically updating the data
  source map repository.  Updates a `domain.DataSourceMapRepository` by using a
  `domain.DataSourceMapRetreiver` to retrieve the latest data source map.
  (NOTE: This code is likely making its way to `domain`).

* `gds` package: contains an implementation of a Ganglia Data Source Server aka
  gmond.  This is the mechanism used to report metrics to Ganglia's gmetad.  The
  metrics reported are read from a `domain.ReportedRepository`.

* `inmem` package: contains in-memory implementations of the repository
  interfaces defined in the `domain` package.

* `processing` package: contains routines for processing the metrics retrieved
  from gmetad.  The processed metrics are stored in an implementation of
  `domain.ProcessedRepository`.  (This code is likely making its way to
  `domain`).

* `repository/memory` package: contains an implementation of
  `domain.Repository` that holds the reported metrics in memory.  (This code is
  likely making its way to `inmem`).

* `retrieval` package: contains routines for periodically retrieving and
  filtering hosts and metrics from ganglia.  The hosts and metrics are parsed
  to `domain.ProcessedHost` and `domain.ProcessedMetric` and made available for
  further processing.

* `ticker` package: contains an wrapper around a `time.Ticker` which also
  allows for manually ticking the ticker.

* `visualizer` package: contains a HTTP client for interacting with the
  Concertim Visualisation App's API.

### History

In legacy and old-legacy Concertim, various daemons, e.g,. `martha`, would
collect metrics from various sources.  Those metrics were made available to
Ganglia's Gmetad by having those daemons listen to TCP connections and respond
with XML.

Gmetad would 1) expire old metrics; 2) aggregate the metrics it collected from
the Concertim daemons; and 3) create and update RRDtool archive files for each
metric.

Another Concertim daemon, `meryl`, would request all of the metrics from Gmetad
and make various views of them available to the rest of Concertim. Those views
included:

1. A list of unique metric names.
2. For each metric, a list of devices reporting that metric.
3. For each device, a list of metrics it is reporting.

In this iteration of Concertim, the control has been inverted.  Concertim no
longer collects metrics, instead they are reported to it by third-party
programs, e.g., `openstack-concertim-service`.

`ct-metric-reporting-daemon` is responsible for the following:

1. Receive metric reports from 3rd-party apps.
2. Report metric to Gmetad.
3. Request all metrics from Gmetad.
4. Process metrics to produce required views.

This current architecture has been arived at through an iterative process
converting from legacy.  It shouldn't be viewed as finished or ideal.  Gmetad
is still involved as it expires metrics and creates/updates the RRDtool archive
files.  Both of those could be moved into `ct-metric-reporting-daemon` but that
has not yet happened.

### Data source maps

Data source maps are used to map between Gmetad's identifier for a host and
Concertim's identifier.  Data source maps are created in `ct-visualisation-app`
and made available to `ct-metrics` via memcache.

Gmetad's identifier for a host is "grid name", "cluster name", "host name"
triple.  Parts of legacy concertim assumed that we would only be interested in
a single grid and cluster.  Parts of new Concertim have carried that assumption
over. This assumption is currently one of the limiting factors requiring that
devices have unique names.


## Starting the server

The server can be started using [air](https://github.com/cosmtrek/air) which
provides automatic reloading for Go apps. `air` is installed as part of the
`appliance-dev` ansible role on the Vagrant machine.  Or to install on your
laptop follow the link above.

## Running tests

The tests can be ran on the vagrant machine or alternatively on your laptop
with the following command:

- All tests:
  ```bash
  go test ./...
  ```

- Tests for a single package, e.g., "api", "gds", etc..
  ```bash
  go test ./api
  ```

- Run all tests on code changes.
  ```bash
  air -build.bin "go test ./..."
  ```

## Developing a feature

1. Select available story in Pivotal tracker.
2. Implement it with test coverage on a feature branch.
3. Create PR https://github.com/alces-flight/ct-metric-reporting-daemon/pulls.
4. Update Pivotal tracker with PR details and mark story as finished.
