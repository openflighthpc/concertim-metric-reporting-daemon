//go:generate go-enum --marshal --lower --names

// Package domain exports the domain models.
//
// It consists of the Cluster, Host and Metric models as well as a Repository
// interface.
package domain

import (
	"fmt"
	"math"
	"reflect"
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

// HostId exists to document some function signatures.
type HostId string

// String implements the Stringer interface.
func (m HostId) String() string {
	return string(m)
}

// DSM represents a Ganglia identifier for a host.
type DSM struct {
	GridName    string
	ClusterName string
	HostName    string
}

// String implements the Stringer interface.
func (d DSM) String() string {
	return fmt.Sprintf("%s/%s/%s", d.GridName, d.ClusterName, d.HostName)
}

// MemcacheKey exists to document some function signatures.
type MemcacheKey string

// String implements the Stringer interface.
func (m MemcacheKey) String() string {
	return string(m)
}

// ReportedHost is the domain model representing a host for which metrics have
// been reported.
type ReportedHost struct {
	Id       HostId
	DSM      DSM
	Reported time.Time
	DMax     time.Duration
	Metrics  []ReportedMetric
}

// MetricName exists to document some function signatures.
type MetricName string

// ReportedMetric is the domain model representing a single reported metric.
// It has not yet been fully processed.
type ReportedMetric struct {
	Name     string
	Value    string
	Units    string
	Slope    MetricSlope
	Reported time.Time
	DMax     time.Duration
	Type     MetricType
}

// ProcessedMetric is the domain model representing a single metric that has
// been fully processed.
type ProcessedMetric struct {
	Name      string
	Datatype  string
	Units     string
	Source    string
	Value     string
	Nature    string
	Dmax      int
	Timestamp int64
	Stale     bool
}

type UniqueMetric struct {
	// XXX Min and max are calculated for all devices across all clusters
	// across all projects.  This is consistent with the initial behaviour
	// of Concertim, but may not be suitable anymore.
	Datatype  string
	Max       any
	Min       any
	Name      string
	Nature    string
	Units     string
}

// ErrInvalidMetricVal is used if the metric's value is not valid for its
// type.
var ErrInvalidMetricVal = fmt.Errorf("not a valid metric value")

// ParseMetricVal attempts to parse the given value according to the given
// metric type.  If successful, the value will be returned as a string.
func ParseMetricVal(val any, metricType MetricType) (string, error) {
	value := reflect.ValueOf(val)
	const epsilon = 1e-9 // Margin of error for converting floats to ints.

	switch metricType {
	case MetricTypeString:
		if value.Kind() == reflect.String {
			return fmt.Sprint(value), nil
		}
	case MetricTypeFloat, MetricTypeDouble:
		if value.CanFloat() {
			return fmt.Sprintf("%f", value.Float()), nil
		}
	case MetricTypeInt8, MetricTypeInt16, MetricTypeInt32:
		if !value.CanFloat() {
			break
		}
		intPart, frac := math.Modf(math.Abs(value.Float()))
		if math.Abs(frac) < epsilon {
			return fmt.Sprintf("%d", int(intPart)), nil
		}
	case MetricTypeUint8, MetricTypeUint16, MetricTypeUint32:
		if !value.CanFloat() || math.Signbit(value.Float()) {
			break
		}
		intPart, frac := math.Modf(math.Abs(value.Float()))
		if math.Abs(frac) < epsilon {
			return fmt.Sprintf("%d", int(intPart)), nil
		}
	default:
		return "", fmt.Errorf("%s is %w", metricType, ErrInvalidMetricType)

	}
	return "", fmt.Errorf("%s is %w", val, ErrInvalidMetricVal)
}
