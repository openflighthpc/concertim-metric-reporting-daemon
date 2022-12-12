//go:generate go-enum --marshal --lower --names

package domain

import (
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

type Cluster struct {
	Hosts []Host
}

type Host struct {
	// XXX Consider Name HostName for better documentation of what we key maps
	// off of.
	Name     string
	Reported time.Time
	TMax     time.Duration
	DMax     time.Duration
	Metrics  []Metric
}

type Metric struct {
	Name   string
	Val    string
	Units  string
	Slope  MetricSlope
	Tn     time.Duration
	TMax   time.Duration
	DMax   time.Duration
	Source string
	Type   MetricType
}
