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

var NumericMetricTypes = []string{"int8", "uint8", "int16", "uint16", "int32", "uint32", "float", "double"}

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

// PendingHost is the domain model representing a host for which metrics have
// been reported.
type PendingHost struct {
	// The Concertim ID for the host.
	Id HostId
	// The data source map for the host.
	DSM DSM
	// The time at which metrics were most recently reported for this host.
	// XXX ProcessedHost has this as a pointer value.
	Reported time.Time
	// A map from metric name to the most recently reported metric with that
	// name.
	Metrics map[MetricName]PendingMetric
}

// MetricName exists to document some function signatures.
type MetricName string

// PendingMetric is the domain model representing a single reported metric.
// It has not yet been fully processed.
type PendingMetric struct {
	Name          string
	Value         string
	Units         string
	Slope         MetricSlope
	Reported      time.Time
	TTL           time.Duration
	Type          MetricType
	LastProcessed *time.Time
}

// ProcessedHost is the domain model representing a single host that has been
// fully processed.
type ProcessedHost struct {
	// The Concertim ID for the host.
	Id      HostId
	DSM     DSM
	Metrics map[MetricName]ProcessedMetric
	// Time that metrics were last reported for the host.
	Mtime *time.Time
}

// ProcessedMetric is the domain model representing a single metric that has
// been fully processed.
type ProcessedMetric struct {
	// XXX Consider changing some of these strings to MetricName etc..
	Name     string
	Datatype string
	Units    string
	Value    string
	Nature   string
	Dmax     int
	// The processing time for the metric.
	Timestamp time.Time
	// Whether the metric has expired.
	Stale bool
}

type UniqueMetric struct {
	// XXX Min and max are calculated for all devices across all clusters
	// across all projects.  This is consistent with the initial behaviour
	// of Concertim, but may not be suitable anymore.
	Datatype string
	Max      any
	Min      any
	Name     string
	Nature   string
	Units    string
}

// HistoricHost is the domain model representing a single host loaded with its
// historic metric values.
type HistoricHost struct {
	// The Concertim ID for the host.
	Id      HostId
	DSM     DSM
	Metrics map[MetricName][]*HistoricMetric
}

type HistoricMetric struct {
	Value     float64
	Timestamp int64
}

type MetricSummary struct {
	// The number of hosts that reported the metric.
	Num int
	// The total value reported for all hosts that report the metric.
	Sum any
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

// HistoricMetricDuration specifies a duration and resolution for
// retrieving common historic metric sets.  E.g., last hour, last day, etc..
type HistoricMetricDuration struct {
	Start      string
	End        string
	Resolution string
}

// LastDuration describes a pre-defined duration for which metrics can be
// retrieved.  E.g., last hour or last day.
//
// NOTE: When adding an entry to ENUM a corresponding entry must be present in
// LastXLookup.
//
// ENUM(hour, day, quarter).
type LastDuration string

var LastXLookup map[LastDuration]HistoricMetricDuration = map[LastDuration]HistoricMetricDuration{
	LastDurationHour:    {Start: "-1h", Resolution: "15s"},
	LastDurationDay:     {Start: "-1d", Resolution: "5m"},
	LastDurationQuarter: {Start: "-90d", Resolution: "1h"},
}

var ErrLastXLookupMissingEntry = fmt.Errorf("missing from lookup map")

func HistoricMetricDurationFromString(duration string) (HistoricMetricDuration, error) {
	lastDuration, err := ParseLastDuration(duration)
	if err != nil {
		return HistoricMetricDuration{}, err
	}
	if x, ok := LastXLookup[lastDuration]; ok {
		return x, nil
	}
	return HistoricMetricDuration{}, fmt.Errorf("%s is %w", duration, ErrLastXLookupMissingEntry)
}

func HistoricMetricDurationFromTimes(startTime, endTime time.Time) HistoricMetricDuration {
	var resolution string
	duration := endTime.Sub(startTime)
	if duration <= 1*time.Hour {
		resolution = "15s"
	} else if duration <= 24*time.Hour {
		resolution = "5m"
	} else {
		resolution = "1h"
	}
	return HistoricMetricDuration{
		Start:      fmt.Sprintf("%d", startTime.Unix()),
		End:        fmt.Sprintf("%d", endTime.Unix()),
		Resolution: resolution,
	}
}
