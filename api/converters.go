package api

import (
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
)

func domainMetricFromPutMetric(src putMetricRequest) (domain.Metric, error) {
	var err error
	var dst domain.Metric
	dst.Name = src.Name
	dst.Units = src.Units
	dst.Reported = time.Now()
	dst.DMax = time.Duration(src.TTL) * time.Second

	dst.Type, err = domain.ParseMetricType(src.Type)
	if err != nil {
		return domain.Metric{}, err
	}
	dst.Val, err = domain.ParseMetricVal(src.Val, dst.Type)
	if err != nil {
		return domain.Metric{}, err
	}
	dst.Slope, err = domain.ParseMetricSlope(src.Slope)
	if err != nil {
		return domain.Metric{}, err
	}
	return dst, nil
}
