package db

import (
	"github.com/alces-flight/concertim-mrapi/domain"
)

type DB interface {
	PutHost(domain.Host) error
	PutMetric(domain.Host, domain.Metric) error
	GetAll() domain.Cluster
}
