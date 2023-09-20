package domain

import (
	"errors"
)

// UnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
var UnknownHost = errors.New("Unknown host")

// ReportedRepository is the interface for storing reported metrics.
type ReportedRepository interface {
	// PutHost adds a Host to the repository.  If the Host has already been
	// added it will be updated.
	PutHost(ReportedHost) error

	// PutMetric adds a Metric to the repository for a previously added Host.
	//
	// If the Metric has already been added it will be updated.
	//
	// If the Host has not been previously added an UnknownHost error is
	// returned.
	PutMetric(ReportedHost, ReportedMetric) error

	// GetAll returns a slice of all Hosts added to the repository, populated
	// with all of their Metrics.
	GetAll() []ReportedHost

	GetHost(HostId) (ReportedHost, bool)
}

// DataSourceMapRepository is the interface for looking up a device's data
// source map to host from its device id.
type DataSourceMapRepository interface {
	// GetDSM returns the data source map for the given host id.
	//
	// deviceId is the device's concertim ID. The returned DSM is the map
	// to that deivce in the Gmetad output.
	GetDSM(deviceId HostId) (dsm DSM, ok bool)

	// GetHostId returns the host id for the given memcache key.
	GetHostId(dsm DSM) (HostId, bool)

	// Update updates the state of the repository.
	Update(map[HostId]DSM, map[DSM]HostId) error
}

type DataSourceMapRetreiver interface {
	GetDSM() (map[HostId]DSM, map[DSM]HostId, error)
}

type DataSourceMapRepoUpdater interface {
	RunPeriodicUpdateLoop()
	UpdateNow()
}

// ProcessedRepository is the interface for storing processed metrics.
type ProcessedRepository interface {
	// GetUniqueMetrics returns a slice of the unique metrics found in the
	// last processing run.  The uniqueness of a metric is determined by
	// its name.
	// XXX GetUniqueMetrics() []*UniqueMetric
	GetUniqueMetrics() []UniqueMetric
	// HostsWithMetric returns a slice of ProcessedHosts that had the given
	// metric in the last processing run.
	HostsWithMetric(metricName MetricName) []*ProcessedHost
	// Begin records the start of a processing run.
	Begin() error
	// Commit commits the results of a processing run.
	Commit() error
	// AddHost records the presence of a host in the current processing run.
	AddHost(host *ProcessedHost)
	// AddMetric records the presence of a metric in the current processing run.
	AddMetric(host *ProcessedHost, metric *ProcessedMetric)
}
