package api

import (
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

func domainMetricFromPutMetric(src putMetricRequest, logger zerolog.Logger) (domain.Metric, error) {
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
		logger.Debug().Err(err).Any("input", src.Val).Stringer("type", dst.Type).Msg("parsing metric failed")
		return domain.Metric{}, err
	}
	logger.Debug().Any("input", src.Val).Stringer("type", dst.Type).Str("result", dst.Val).Msg("converted metric value")
	dst.Slope, err = domain.ParseMetricSlope(src.Slope)
	if err != nil {
		return domain.Metric{}, err
	}
	return dst, nil
}
