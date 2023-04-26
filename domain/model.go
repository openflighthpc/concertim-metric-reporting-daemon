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
