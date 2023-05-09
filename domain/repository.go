package domain

import "errors"

// UnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
var UnknownHost = errors.New("Unknown host")

// Repository is the interface for any persistence layer.
type Repository interface {
	// PutHost adds a Host to the repository.  If the Host has already been
	// added it will be updated.
	PutHost(Host) error

	// PutMetric adds a Metric to the repository for a previously added Host.
	//
	// If the Metric has already been added it will be updated.
	//
	// If the Host has not been previously added an UnknownHost error is
	// returned.
	PutMetric(Host, Metric) error

	// GetAll returns a slice of all Hosts added to the repository, populated
	// with all of their Metrics.
	GetAll() []Host

	GetHost(Hostname) (Host, bool)
}

// DataSourceMapRepository is the interface for looking up a device's data
// source map to host from its device name.
type DataSourceMapRepository interface {
	// GetDSM returns the data source map for the given host name.
	//
	// deviceName is a user-friendly name used in the display of the
	// appliance.  The returned DSM may or may not be user-friendly it
	// is used only internally.  Examples of the mapping are below.
	//
	// * comp001      -> unspecified/unspecified/comp01.concertim.alces-flight.com
	// * tempsensor01 -> unspecified/unspecified/sensor-dd5fb19b50624b33e1c5e4d5003714f4
	// * pdu01        -> unspecified/unspecified/rack_2__powerstrip__startu42__1669827901
	GetDSM(deviceName Hostname) (dsm DSM, ok bool)

	// GetMemcacheKey returns the memcache key for the given data source map.
	GetMemcacheKey(dsm DSM) (MemcacheKey, bool)

	// Update retrieves the latest DSM from an external source and updates
	// its internal repository.
	//
	// The external source to use is configured when creating a new DSMRepo.
	Update() error
}
