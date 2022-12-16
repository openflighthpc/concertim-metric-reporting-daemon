// Code generated by github.com/jmattheis/goverter, DO NOT EDIT.

package memory

import (
	domain "github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"time"
)

type ConverterImpl struct{}

func (c *ConverterImpl) DomainFromModelHost(source HostModel) domain.Host {
	var domainHost domain.Host
	domainHost.DeviceName = source.DeviceName
	domainHost.DSMName = source.DSMName
	domainHost.Reported = ConvertTime(source.Reported)
	domainHost.DMax = time.Duration(source.DMax)
	domainHost.Metrics = DefaultMetrics()
	return domainHost
}
func (c *ConverterImpl) DomainFromModelMetric(source MetricModel) domain.Metric {
	var domainMetric domain.Metric
	domainMetric.Name = source.Name
	domainMetric.Val = source.Val
	domainMetric.Units = source.Units
	domainMetric.Slope = domain.MetricSlope(source.Slope)
	domainMetric.Reported = ConvertTime(source.Reported)
	domainMetric.DMax = time.Duration(source.DMax)
	domainMetric.Type = domain.MetricType(source.Type)
	return domainMetric
}
func (c *ConverterImpl) ModelFromDomainHost(source domain.Host) HostModel {
	var memoryHostModel HostModel
	memoryHostModel.DeviceName = source.DeviceName
	memoryHostModel.DSMName = source.DSMName
	memoryHostModel.Reported = ConvertTime(source.Reported)
	memoryHostModel.DMax = time.Duration(source.DMax)
	return memoryHostModel
}
func (c *ConverterImpl) ModelFromDomainMetric(source domain.Metric) MetricModel {
	var memoryMetricModel MetricModel
	memoryMetricModel.Name = source.Name
	memoryMetricModel.Val = source.Val
	memoryMetricModel.Units = source.Units
	memoryMetricModel.Slope = domain.MetricSlope(source.Slope)
	memoryMetricModel.Reported = ConvertTime(source.Reported)
	memoryMetricModel.DMax = time.Duration(source.DMax)
	memoryMetricModel.Type = domain.MetricType(source.Type)
	return memoryMetricModel
}
