package gds

import (
	"encoding/xml"
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
)

type cluster struct {
	XMLName   xml.Name `xml:"CLUSTER"`
	Name      string   `xml:"NAME,attr"`
	LocalTime int64    `xml:"LOCALTIME,attr"`
	Owner     string   `xml:"OWNER,attr"`
	LatLong   string   `xml:"LATLONG,attr"`
	URL       string   `xml:"URL,attr"`
	Hosts     []host   `xml:"HOSTS"`
}

type host struct {
	XMLName  xml.Name `xml:"HOST"`
	Name     string   `xml:"NAME,attr"`
	IP       string   `xml:"IP,attr"`
	Reported int64    `xml:"REPORTED,attr"`
	Tn       int      `xml:"TN,attr"`
	TMax     int      `xml:"TMAX,attr"`
	DMax     int      `xml:"DMAX,attr"`
	Metrics  []metric `xml:"METRICS"`
}

type metric struct {
	XMLName xml.Name           `xml:"METRIC"`
	Name    string             `xml:"NAME,attr"`
	Val     string             `xml:"VAL,attr"`
	Units   string             `xml:"UNITS,attr"`
	Slope   domain.MetricSlope `xml:"SLOPE,attr"`
	Tn      int                `xml:"TN,attr"`
	TMax    int                `xml:"TMAX,attr"`
	DMax    int                `xml:"DMAX,attr"`
	Source  string             `xml:"SOURCE,attr"`
	Type    domain.MetricType  `xml:"TYPE,attr"`
}

var header = []byte(`<?xml version="1.0" encoding="ISO-8859-1"?>
<GANGLIA_XML VERSION="3.1.7" SOURCE="mrapi">
`)

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type outputGenerator struct {
	clock  clock
	header []byte
}

func newOutputGenerator(clock clock) *outputGenerator {
	return &outputGenerator{
		clock:  clock,
		header: header,
	}
}

func (g *outputGenerator) generate(dCluster domain.Cluster) ([]byte, error) {
	cluster := g.clusterFromDomain(dCluster)
	xml, err := xml.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return nil, err
	}
	xml = append(xml, []byte("\n")...)
	return append(g.header, xml...), nil
}

func (g *outputGenerator) clusterFromDomain(dCluster domain.Cluster) cluster {
	cluster := cluster{
		Name:      "unspecified",
		LocalTime: g.clock.Now().Unix(),
		Owner:     "unspecified",
		LatLong:   "unspecified",
		URL:       "unspecified",
		// XXX
		// Hosts:     hosts,
	}

	for _, dh := range dCluster.Hosts {
		cluster.Hosts = append(cluster.Hosts, g.hostFromDomain(dh))
	}

	return cluster
}

func (g *outputGenerator) hostFromDomain(dHost domain.Host) host {
	host := host{
		Name:     dHost.Name,
		IP:       "",
		Reported: dHost.Reported.Unix(),
		Tn:       0,
		TMax:     int(dHost.TMax.Seconds()),
		DMax:     int(dHost.DMax.Seconds()),
	}
	// Metrics: []metric{},
	for _, dm := range dHost.Metrics {
		host.Metrics = append(host.Metrics, g.metricsFromDomain(dm))
	}
	return host
}

func (g *outputGenerator) metricsFromDomain(dMetric domain.Metric) metric {
	return metric{
		Name:   dMetric.Name,
		Val:    dMetric.Val,
		Units:  dMetric.Units,
		Slope:  dMetric.Slope,
		Tn:     int(dMetric.Tn.Seconds()),
		TMax:   int(dMetric.TMax.Seconds()),
		DMax:   int(dMetric.DMax.Seconds()),
		Source: dMetric.Source,
		Type:   dMetric.Type,
	}
}
