package domain

import "fmt"

// UnknownHost is the error reported when an attempt to add a metric to an
// unknown host is made.
type UnknownHost struct {
	Host Host
}

func (e UnknownHost) Error() string {
	return fmt.Sprintf("Unknown host %s", e.Host.Name)
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

	// GetAll returns a Cluster populated with all of the Hosts and Metrics
	// that have been added to the repository.
	GetAll() Cluster
}
