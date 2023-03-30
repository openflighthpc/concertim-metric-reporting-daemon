//go:generate go-enum --marshal --lower --names

package retrieval

import "encoding/xml"

type gangliaRoot struct {
	XMLName xml.Name `xml:"GANGLIA_XML"`
	VERSION string   `xml:"VERSION,attr"`
	Grids   []Grid   `xml:"GRID"`
}

type Grid struct {
	XMLName  xml.Name  `xml:"GRID"`
	Name     string    `xml:"NAME,attr"`
	Clusters []Cluster `xml:"CLUSTER"`
}

type Cluster struct {
	XMLName xml.Name `xml:"CLUSTER"`
	Name    string   `xml:"NAME,attr"`
	Hosts   []Host   `xml:"HOST"`
}

type Host struct {
	// DSM     DSM
	Name    string   `xml:"NAME,attr"`
	Metrics []Metric `xml:"METRIC"`
}

// Metric is the domain model representing a single metric.
type Metric struct {
	Name   string `xml:"NAME,attr"`
	Val    string `xml:"VAL,attr"`
	Units  string `xml:"UNITS,attr"`
	Slope  string `xml:"SLOPE,attr"`
	DMax   string `xml:"DMAX,attr"`
	TMax   string `xml:"TMAX,attr"`
	TN     string `xml:"TN,attr"`
	Type   string `xml:"TYPE,attr"`
	Source string `xml:"SOURCE,attr"`
}

// // MetricSlope describes how the value of the metric can change overtime.
// //
// // MetricSlopeZero: values are not comaprible, e.g., operating system name.
// // MetricSlopePositive: value only increases over time, e.g., total downloads.
// // MetricSlopeNegative: value only decreases over time.
// // MetricSlopeBoth: value can increase or decrease over time.
// // MetricSlopeDerivative: XXX What is this for?
// //
// // ENUM(zero, positive, negative, both, derivative).
// type MetricSlope string
//
// // MetricType describes the data type of the metric.
// //
// // ENUM(string, int8, uint8, int16, uint16, int32, uint32, float, double).
// type MetricType string
