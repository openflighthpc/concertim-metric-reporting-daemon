package domain

import "fmt"

type UnknownHost struct {
	Host Host
}

func (e UnknownHost) Error() string {
	return fmt.Sprintf("Unknown host %s", e.Host.Name)
}

type Repository interface {
	PutHost(Host) error
	PutMetric(Host, Metric) error
	GetAll() Cluster
}
