package api

import (
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

func domainMetricFromPutMetric(src putMetricRequest, logger zerolog.Logger) (domain.ReportedMetric, error) {
	var err error
	var dst domain.ReportedMetric
	dst.Name = src.Name
	dst.Units = src.Units
	dst.Reported = time.Now()
	dst.DMax = time.Duration(src.TTL) * time.Second

	dst.Type, err = domain.ParseMetricType(src.Type)
	if err != nil {
		return domain.ReportedMetric{}, err
	}
	dst.Value, err = domain.ParseMetricVal(src.Val, dst.Type)
	if err != nil {
		logger.Debug().Err(err).Any("input", src.Val).Stringer("type", dst.Type).Msg("parsing metric failed")
		return domain.ReportedMetric{}, err
	}
	logger.Debug().Any("input", src.Val).Stringer("type", dst.Type).Str("result", dst.Value).Msg("converted metric value")
	dst.Slope, err = domain.ParseMetricSlope(src.Slope)
	if err != nil {
		return domain.ReportedMetric{}, err
	}
	return dst, nil
}
