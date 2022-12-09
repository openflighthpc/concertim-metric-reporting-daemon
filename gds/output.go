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

func generateOutput(dCluster domain.Cluster) ([]byte, error) {
	cluster := gdsClusterFromDomain(dCluster)
	xml, err := xml.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return nil, err
	}
	return xml, nil
}

func gdsClusterFromDomain(dCluster domain.Cluster) cluster {
	cluster := cluster{
		Name:      "unspecified",
		LocalTime: time.Now().Unix(),
		Owner:     "unspecified",
		LatLong:   "unspecified",
		URL:       "unspecified",
		// XXX
		// Hosts:     hosts,
	}

	for _, dh := range dCluster.Hosts {
		cluster.Hosts = append(cluster.Hosts, gdsHostFromDomain(dh))
	}

	return cluster
}

func gdsHostFromDomain(dHost domain.Host) host {
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
		host.Metrics = append(host.Metrics, gdsMetricFromDomain(dm))
	}
	return host
}

func gdsMetricFromDomain(dMetric domain.Metric) metric {
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
