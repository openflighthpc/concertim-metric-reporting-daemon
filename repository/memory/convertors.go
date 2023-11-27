// Code generated by github.com/jmattheis/goverter, DO NOT EDIT.

package memory

import (
	domain "github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"time"
)

type ConverterImpl struct{}

func (c *ConverterImpl) DomainFromModelHost(source HostModel) domain.PendingHost {
	var domainReportedHost domain.PendingHost
	domainReportedHost.Id = domain.HostId(source.DeviceId)
	domainReportedHost.DSM = DSMNameToDSM(source.DSMName)
	domainReportedHost.Reported = ConvertTime(source.Reported)
	domainReportedHost.DMax = time.Duration(source.DMax)
	domainReportedHost.Metrics = DefaultMetrics()
	return domainReportedHost
}
func (c *ConverterImpl) DomainFromModelMetric(source MetricModel) domain.PendingMetric {
	var domainReportedMetric domain.PendingMetric
	domainReportedMetric.Name = source.Name
	domainReportedMetric.Value = source.Val
	domainReportedMetric.Units = source.Units
	domainReportedMetric.Slope = domain.MetricSlope(source.Slope)
	domainReportedMetric.Reported = ConvertTime(source.Reported)
	domainReportedMetric.TTL = time.Duration(source.DMax)
	domainReportedMetric.Type = domain.MetricType(source.Type)
	return domainReportedMetric
}
func (c *ConverterImpl) ModelFromDomainHost(source domain.PendingHost) HostModel {
	var memoryHostModel HostModel
	memoryHostModel.DeviceId = string(source.Id)
	memoryHostModel.DSMName = source.DSM.HostName
	memoryHostModel.Reported = ConvertTime(source.Reported)
	memoryHostModel.DMax = time.Duration(source.DMax)
	return memoryHostModel
}
func (c *ConverterImpl) ModelFromDomainMetric(source domain.PendingMetric) MetricModel {
	var memoryMetricModel MetricModel
	memoryMetricModel.Name = source.Name
	memoryMetricModel.Val = source.Value
	memoryMetricModel.Units = source.Units
	memoryMetricModel.Slope = domain.MetricSlope(source.Slope)
	memoryMetricModel.Reported = ConvertTime(source.Reported)
	memoryMetricModel.DMax = time.Duration(source.TTL)
	memoryMetricModel.Type = domain.MetricType(source.Type)
	return memoryMetricModel
}
