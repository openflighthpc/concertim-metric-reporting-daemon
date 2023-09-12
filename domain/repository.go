package domain

import (
	"errors"
)

// UnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
var UnknownHost = errors.New("Unknown host")

// Repository is the interface for any persistence layer.
type Repository interface {
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
	// deviceId is the device's concertim ID. The returned DSM may or may
	// not be user-friendly it is used only internally.
	// mapping are below.
	//
	// * 1 -> unspecified/unspecified/hacor:device:1
	GetDSM(deviceId HostId) (dsm DSM, ok bool)

	// GetMemcacheKey returns the memcache key for the given data source map.
	GetMemcacheKey(dsm DSM) (MemcacheKey, bool)

	// Update retrieves the latest DSM from an external source and updates
	// its internal repository.
	//
	// The external source to use is configured when creating a new DSMRepo.
	Update() error
}

type ResultRepo interface {
	// GetHostMetrics(deviceId HostId) (metrics map[MetricName]Metric, ok bool)
	GetUniqueMetrics() []UniqueMetric
}
