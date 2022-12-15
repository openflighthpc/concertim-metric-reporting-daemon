//go:generate go-enum --marshal --lower --names

// Package domain exports the domain models.
//
// It consists of the Cluster, Host and Metric models as well as a Repository
// interface.
package domain

import (
	"fmt"
	"time"
)

// MetricSlope describes how the value of the metric can change overtime.
//
// MetricSlopeZero: values are not comaprible, e.g., operating system name.
// MetricSlopePositive: value only increases over time, e.g., total downloads.
// MetricSlopeNegative: value only decreases over time.
// MetricSlopeBoth: value can increase or decrease over time.
// MetricSlopeDerivative: XXX What is this for?
//
// ENUM(zero, positive, negative, both, derivative).
type MetricSlope string

// MetricType describes the data type of the metric.
//
// ENUM(string, int8, uint8, int16, uint16, int32, uint32, float, double).
type MetricType string

// Cluster is the domain model representing a cluster.
type Cluster struct {
	Hosts []Host
}

// Host is the domain model representing a host.
type Host struct {
	// XXX Consider Name HostName for better documentation of what we key maps
	// off of.
	DeviceName string
	DSMName    string
	Reported   time.Time
	DMax       time.Duration
	Metrics    []Metric
}

// Metric is the domain model representing a single metric.
type Metric struct {
	Name     string
	Val      string
	Units    string
	Slope    MetricSlope
	Reported time.Time
	DMax     time.Duration
	Type     MetricType
}

var ErrInvalidMetricVal = fmt.Errorf("not a valid metric value")

func ParseMetricVal(val any, metricType MetricType) (string, error) {
	switch v := val.(type) {
	case string:
		if metricType == MetricTypeString {
			return v, nil
		}
		if v, ok := val.(float64); ok {
			return "", fmt.Errorf("%f is %w sbe glcr %s", v, ErrInvalidMetricVal, metricType)
		}
		return "", fmt.Errorf("%s is %w for type %s", val, ErrInvalidMetricVal, metricType)
	case float64:
		if metricType == MetricTypeString {
			return "", fmt.Errorf("%f is %w for type %s", val, ErrInvalidMetricVal, metricType)
		}
		return parseFloat64ToMetricType(v, metricType)
	default:
		return "", fmt.Errorf("%s is %w", val, ErrInvalidMetricVal)
	}
}

func parseFloat64ToMetricType(val float64, metricType MetricType) (string, error) {
	switch metricType {
	case MetricTypeInt8,
		MetricTypeUint8,
		MetricTypeInt16,
		MetricTypeUint16,
		MetricTypeInt32,
		MetricTypeUint32:
		return fmt.Sprintf("%d", int(val)), nil
	case MetricTypeFloat, MetricTypeDouble:
		return fmt.Sprintf("%f", val), nil
	}
	return "", nil
}
