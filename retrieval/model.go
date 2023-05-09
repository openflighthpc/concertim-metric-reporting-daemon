package retrieval

import "encoding/xml"

type gangliaRoot struct {
	XMLName xml.Name `xml:"GANGLIA_XML"`
	VERSION string   `xml:"VERSION,attr"`
	Grids   []Grid   `xml:"GRID"`
}

// Grid represents a single ganglia grid.
type Grid struct {
	XMLName  xml.Name  `xml:"GRID"`
	Name     string    `xml:"NAME,attr"`
	Clusters []Cluster `xml:"CLUSTER"`
}

// Cluster represents a single ganglia cluster.
type Cluster struct {
	XMLName xml.Name `xml:"CLUSTER"`
	Name    string   `xml:"NAME,attr"`
	Hosts   []Host   `xml:"HOST"`
}

// Host represents a single ganglia host.
type Host struct {
	Name    string   `xml:"NAME,attr"`
	Metrics []Metric `xml:"METRIC"`
}

// Metric represents a single ganglia metric.
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
