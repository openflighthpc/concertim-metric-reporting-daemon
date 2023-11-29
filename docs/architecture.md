# Metric Reporting Daemon Architecture

`metric-reporting-daemon` is written in Go and provides an HTTP API for metrics
to be reported to Concertim and to retrieve details about those metrics.

## Metric processing, pending, current and historic repositories

MRD contains three different repositories for the metrics that are reported to
it: a "pending repository"; a "current repository" and a "historic repository".

When a metric is reported to MRD it is added to the pending repository.
Periodically, MRD fetches all of the metrics in the pending repository, (1)
calculates summaries of them; (2) updates the historic repository with those
metrics; and (3) sets the current repository to the processed metrics.

The HTTP API for querying metrics allows for querying both the current and the
historic metrics.  The current metrics are those that were processed in the
last processing run.  The historic metrics are those that have been processed
in any processing run.  The pending metrics are not available for querying.

The current repository contains two views of the most recently processed metrics.

1. a list of unique metrics (where metric uniqueness is determined by the metric's name).
2. for each metric a list of devices that reported that metric in the last processing run.

The historic repository contains the following

1. a list of all metric names that have been reported.
2. for each device, a list of all metric names that have been reported for that device.
3. for each device and metric, the historic values that have been reported.

### Metric expiration

When a metric is reported, a time-to-live (TTL) is given for it.  The metrics
expiration time is calculated from the time it was reported and its TTL.  That
metric continues to be processed with the reported value until:

1. A new value is reported for the metric.
2. The metric's expiration time is reached.

This means that an infrequently changing metric does not need to be reported
with the same frequency as a frequently changing metric.

### Metric aggregation and summaries

For each device and metric, the following aggregations are used when storing
the historic metrics:

* The average, minimum and maximum values, calcualted over 15 seconds are
stored for the most recent hour.
* The average, minimum and maximum values, calcualted over 5 minutes are
stored for the most recent day.
* The average, minimum and maximum values, calcualted over 1 hour are
stored for the most 90 days.

Metrics that are over 90 days old are not stored.

For each device:metric pair, the value reported is the value processed by the
above aggregations.

For each metric, summary containing the number of devices reporting the metric
and the sum of the values reported is calculated.  This summary is stored using
the same aggregations.

It is expected that eventually we will want to produce summaries for additional
groups of devices, such as all devices in a rack, or all devices belonging to a
user.  For now we calculate a single summary for all devices regardless of
rack, cluster or project.

## Directories

A brief explanation of the directories is as follows:

* `api` contains the HTTP API server for reporting and querying
   metrics.  Reported metrics are stored in an implementation of
  `domain.PendingRepository`.  Metrics are queried from an implementation of
  `domain.CurrentRepository`.

* `canned` directory contains functionality to provide canned responses for
  querying the data source maps.  This allows for developing MRD without a
  running `ct-visualisation-app`.

* `cmd` directory contains the executables.

* `config` contains the config files and config related code.

* `docker` contains files for building the docker image.

* `domain` contains the domain entities.  Currently an `Application`
  struct, various models to represent both pending and current hosts
  and metrics, and various repository interfaces.  It also contains the code
  for processing the pending metrics to become current metrics.

* `dsmRepository` contains code for periodically updating the data
  source map repository.  Updates a `domain.DataSourceMapRepository` by using a
  `domain.DataSourceMapRetreiver` to retrieve the latest data source map.
  (NOTE: This code is likely making its way to `domain`).

* `inmem` contains in-memory implementations of the repository interfaces
  defined in the `domain` package.

* `rrd` contains an implementation of the `domain.HistoricRepository` interface
  for storing and retrieving historic metrics using RRDTool.

* `visualizer` contains a HTTP client for interacting with the Concertim
  Visualisation App's API.

## History

The architecture of MRD is still somewhat influenced by legacy and old-legacy
Concertim.  This influence should be removed in time.

That influence is still present in the use of data source maps and in the
particular parameters expected when a metric is reported, e.g., support for
data types `int8`, `uint8` etc. and specification of a `slope. 

It is expected that this influence will be lessened and even removed in future.

### Data source maps

Data source maps were originally used to map between Ganglia's Gmetad's
identifier for a host and Concertim's identifier.  Data source maps are created
in `ct-visualisation-app` and made available to `ct-metric-reporting-daemon`
via `ct-visualisation-app`'s HTTP API.

Gmetad's identifier for a host is "grid name", "cluster name", "host name"
triple.  Parts of legacy concertim assumed that we would only be interested in
a single grid and cluster.  Parts of new Concertim have carried that assumption
over.


## Starting the server

See [/docs/DEVELOPMENT.md](/docs/DEVELOPMENT.md) for details on how to start the server.
