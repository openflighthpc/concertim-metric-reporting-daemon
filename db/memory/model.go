package memory

import (
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
)

type Host struct {
	Name     string
	Reported time.Time
	TMax     time.Duration
	DMax     time.Duration
}

type Metric struct {
	Name   string
	Val    string
	Units  string
	Slope  domain.MetricSlope
	Tn     time.Duration
	TMax   time.Duration
	DMax   time.Duration
	Source string
	Type   domain.MetricType
}

func dbHostFromDomain(dh domain.Host) Host {
	return Host{
		Name:     dh.Name,
		Reported: dh.Reported,
		TMax:     dh.TMax,
		DMax:     dh.DMax,
	}
}

func dbMetricFromDomain(dm domain.Metric) Metric {
	return Metric{
		Name:  dm.Name,
		Val:   dm.Val,
		Units: dm.Units,
		Slope: dm.Slope,
		Tn:    dm.Tn,
		TMax:  dm.TMax,
		DMax:  dm.DMax,
		Type:  dm.Type,
	}
}

func domainHostFromDb(dh Host) domain.Host {
	return domain.Host{
		Name:     dh.Name,
		Reported: dh.Reported,
		TMax:     dh.TMax,
		DMax:     dh.DMax,
		Metrics:  []domain.Metric{},
	}
}

func domainMetricFromDb(dm Metric) domain.Metric {
	return domain.Metric{
		Name:   dm.Name,
		Val:    dm.Val,
		Units:  dm.Units,
		Slope:  dm.Slope,
		Tn:     dm.Tn,
		TMax:   dm.TMax,
		DMax:   dm.DMax,
		Source: "mrapi",
		Type:   dm.Type,
	}
}
