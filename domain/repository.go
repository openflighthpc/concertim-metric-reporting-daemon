package domain

import "fmt"

// UnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
type UnknownHost struct {
	HostName string
}

func (e UnknownHost) Error() string {
	return fmt.Sprintf("Unknown host %s", e.HostName)
}

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

	GetHost(string) (Host, bool)
}

// DataSourceMapRepository is the interface for looking up a device's data
// source map to host from its device name.
type DataSourceMapRepository interface {
	// Get returns the data source map to host for the given device name.
	//
	// deviceName is a user-friendly name used in the display of the
	// appliance.  The returned mapToHost may or may not be user-friendly it
	// is used only internally.  Examples of the mapping are below.
	//
	// * comp001      -> comp01.concertim.alces-flight.com
	// * tempsensor01 -> sensor-dd5fb19b50624b33e1c5e4d5003714f4
	// * pdu01        -> rack_2__powerstrip__startu42__1669827901
	Get(deviceName string) (mapToHost string, ok bool)
}
