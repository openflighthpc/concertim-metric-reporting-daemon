//go:generate go-enum --marshal --lower --names

// Package domain exports the domain models.
//
// It consists of the Cluster, Host and Metric models as well as a Repository
// interface.
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

// Cluster is the domain model representing a cluster.
type Cluster struct {
	Hosts []Host
}

// Host is the domain model representing a host.
type Host struct {
	// XXX Consider Name HostName for better documentation of what we key maps
	// off of.
	Name     string
	Reported time.Time
	DMax     time.Duration
	Metrics  []Metric
}

// Metric is the domain model representing a single metric.
type Metric struct {
	Name  string
	Val   string
	Units string
	Slope MetricSlope
	Tn    time.Duration
	DMax  time.Duration
	Type  MetricType
}
