package domain

import (
	"errors"
	"time"
)

// ErrUnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
var ErrUnknownHost = errors.New("Unknown host")
var ErrWaitingOnProcessingRun = errors.New("Waiting on metric processing run")
var ErrHostNotFound = errors.New("Host not found")
var ErrMetricNotFound = errors.New("Metric not found")

// PendingRepository is the interface for storing reported metrics that have
// not yet been processed.  Metrics in this repository are processed
// periodically and once processed become the current metrics.
type PendingRepository interface {
	// PutHost adds a Host to the repository.  If the Host has already been
	// added it will be updated.
	PutHost(PendingHost) error

	// PutMetric adds a Metric to the repository for a previously added Host.
	//
	// If the Metric has already been added it will be updated.
	//
	// If the Host has not been previously added an UnknownHost error is
	// returned.
	PutMetric(PendingHost, PendingMetric) error

	// GetAll returns a slice of all Hosts added to the repository, populated
	// with all of their Metrics.
	GetAll() []PendingHost

	// GetHost returns the host identified by HostId if present.
	GetHost(HostId) (PendingHost, bool)

	// UpdateLastProcessed updates the metric's LastProcessed field.
	UpdateLastProcessed(HostId, MetricName, time.Time) error
}

// DataSourceMapRepository is the interface for looking up a device's data
// source map to host from its device id.
type DataSourceMapRepository interface {
	// GetDSM returns the data source map for the given host id.
	//
	// deviceId is the device's concertim ID. The returned DSM is the
	// hierarchical map to that device's RRD files.
	GetDSM(deviceId HostId) (dsm DSM, ok bool)

	// GetHostId returns the host id for the given data source map.
	GetHostId(dsm DSM) (HostId, bool)

	// Update updates the state of the repository.
	Update(map[HostId]DSM, map[DSM]HostId) error
}

// DataSourceMapRetreiver is the interface for retrieving the latest data
// source map.
type DataSourceMapRetreiver interface {
	GetDSM() (map[HostId]DSM, map[DSM]HostId, error)
}

// DataSourceMapRetreiver is the interface for updating a DataSourceMapRepository.
type DataSourceMapRepoUpdater interface {
	RunPeriodicUpdateLoop()
	UpdateNow()
}

// CurrentRepository is the interface for storing the most recently processed
// metrics.
type CurrentRepository interface {
	// GetUniqueMetrics returns a slice of the unique metrics found in the
	// last processing run.  The uniqueness of a metric is determined by
	// its name.
	GetUniqueMetrics() ([]*UniqueMetric, error)
	// GetMetricsForHost returns the metrics reported by the given host in the
	// most recent processing run.
	GetMetricsForHost(hostId HostId) ([]*CurrentMetric, error)
	// HostsWithMetric returns a slice of CurrentHosts that had the given
	// metric in the last processing run.
	HostsWithMetric(metricName MetricName) ([]*CurrentHost, error)
	// Begin records the start of a processing run.
	Begin() error
	// Commit commits the results of a processing run.
	Commit() error
	// AddHost records the presence of a host in the current processing run.
	AddHost(host *CurrentHost)
	// AddMetric records the presence of a metric in the current processing run.
	AddMetric(host *CurrentHost, metric *CurrentMetric)
}

// HistoricRepository is the interface for storing and retrieving historic
// metrics.
type HistoricRepository interface {
	// GetValuesForMetric returns all historic values for all hosts that
	// reported the metric in the given duration.
	GetValuesForMetric(metricName MetricName, lastConfig HistoricMetricDuration) ([]*HistoricHost, error)
	// GetValuesForHostAndMetric returns all historic values for the given host
	// and metric between the given duration.
	GetValuesForHostAndMetric(hostId HostId, metricName MetricName, lastConfig HistoricMetricDuration) (*HistoricHost, error)
	// ListMetricNames lists all historic metric names for all hosts.  If a
	// metric is reported for more than one host it will only be included once.
	ListMetricNames() ([]string, error)
	// ListHostMetricNames lists all historic metric names for the given hosts.
	ListHostMetricNames(hostId HostId) ([]string, error)
	// UpdateHostMetric updates the historic record for the given host and
	// metric with the metric's current value.
	UpdateMetric(host *CurrentHost, metric *CurrentMetric) error
	// UpdateSummaryMetrics updates the historic record for the given
	// summaries.
	UpdateSummaryMetrics(MetricSummaries) error
}

// MetricSummaries is the interface for calculating metric summaries.  The
// summaries are calculated as part of the periodic processing run.  Once
// calculated they have to be persisted by calling
// HistoricRepository.UpdateSummaryMetrics.
type MetricSummaries interface {
	// AddMetric updates the summary of the metric.
	AddMetric(metric CurrentMetric) error
	// GetSummaries returns a map of metric name to metric summary.
	GetSummaries() map[MetricName]*MetricSummary
}
