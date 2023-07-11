//go:generate goverter -packageName=memory  -packagePath=github.com/alces-flight/concertim-metric-reporting-daemon/repository/memory -output ./convertors.go .

package memory

import (
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
)

var conv Converter = &ConverterImpl{}

// HostModel is the format of a host as stored in this repository.
type HostModel struct {
	DeviceId string
	DSMName  string
	Reported time.Time
	DMax     time.Duration
}

// MetricModel is the format of a metric as stored in this repository.
type MetricModel struct {
	Name     string
	Val      string
	Units    string
	Slope    domain.MetricSlope
	Reported time.Time
	DMax     time.Duration
	Type     domain.MetricType
}

// Converter is an interface used to convert from domain models to repositry
// models.  goverter is used to automatically generate methods to do so.
//
// goverter:converter
// goverter:extend ConvertTime
// goverter:extend DSMNameToDSM
type Converter interface {
	// goverter:map Id DeviceId
	// goverter:map DSM.HostName DSMName
	ModelFromDomainHost(source domain.Host) HostModel
	ModelFromDomainMetric(source domain.Metric) MetricModel
	DomainFromModelMetric(source MetricModel) domain.Metric
	// goverter:map DeviceId Id
	// goverter:map DSMName DSM
	// goverter:mapExtend Metrics DefaultMetrics
	DomainFromModelHost(source HostModel) domain.Host
}

func domainHostFromModelHostAndMetrics(modelHost HostModel, modelMetrics map[string]MetricModel) domain.Host {
	domainHost := conv.DomainFromModelHost(modelHost)
	for _, modelMetric := range modelMetrics {
		domainHost.Metrics = append(domainHost.Metrics, conv.DomainFromModelMetric(modelMetric))
	}

	return domainHost
}

// ConvertTime converts from a time.Time to a time.Time.
func ConvertTime(source time.Time) time.Time {
	return source
}

// DefaultMetrics sets the default metrics for domain.Host when retrieved from
// the repository.
func DefaultMetrics() []domain.Metric {
	return make([]domain.Metric, 0)
}

// DSMNameToDSM builds a domain.DSM from a HostModel.DSMName.
func DSMNameToDSM(hName string) domain.DSM {
	dsm := domain.DSM{
		GridName:    "unspecified",
		ClusterName: "unspecified",
		HostName:    hName,
	}
	return dsm
}
