package domain

type Repository interface {
	PutHost(Host) error
	PutMetric(Host, Metric) error
	GetAll() Cluster
}
