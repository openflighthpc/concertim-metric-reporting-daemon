//go:generate goverter -packageName=memory  -packagePath=github.com/alces-flight/concertim-mrapi/db/memory -output ./convertors.go .

package memory

import (
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
)

var conv Converter = &ConverterImpl{}

type HostModel struct {
	Name     string
	Reported time.Time
	TMax     time.Duration
	DMax     time.Duration
}

type MetricModel struct {
	Name   string
	Val    string
	Units  string
	Slope  domain.MetricSlope
	Tn     time.Duration
	TMax   time.Duration
	DMax   time.Duration
	Source string
	Type   domain.MetricType
}

// goverter:converter
// goverter:extend ConvertTime
type Converter interface {
	ModelFromDomainHost(source domain.Host) HostModel
	ModelFromDomainMetric(source domain.Metric) MetricModel
	DomainFromModelMetric(source MetricModel) domain.Metric
	// goverter:mapExtend Metrics DefaultMetrics
	DomainFromModelHost(source HostModel) domain.Host
}

func DomainHostFromModelHostAndMetrics(modelHost HostModel, modelMetrics map[string]MetricModel) domain.Host {
	domainHost := conv.DomainFromModelHost(modelHost)
	for _, modelMetric := range modelMetrics {
		domainHost.Metrics = append(domainHost.Metrics, conv.DomainFromModelMetric(modelMetric))
	}

	return domainHost
}

func ConvertTime(source time.Time) time.Time {
	return source
}

func DefaultMetrics() []domain.Metric {
	return make([]domain.Metric, 0)
}
