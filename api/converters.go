package api

import (
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

func domainMetricFromPutMetric(src putMetricRequest, logger zerolog.Logger) (domain.PendingMetric, error) {
	var err error
	var dst domain.PendingMetric
	dst.Name = src.Name
	dst.Units = src.Units
	dst.Reported = time.Now()
	dst.TTL = time.Duration(src.TTL) * time.Second

	dst.Type, err = domain.ParseMetricType(src.Type)
	if err != nil {
		return domain.PendingMetric{}, err
	}
	dst.Value, err = domain.ParseMetricVal(src.Val, dst.Type)
	if err != nil {
		logger.Debug().Err(err).Any("input", src.Val).Stringer("type", dst.Type).Msg("parsing metric failed")
		return domain.PendingMetric{}, err
	}
	logger.Debug().Any("input", src.Val).Stringer("type", dst.Type).Str("result", dst.Value).Msg("converted metric value")
	dst.Slope, err = domain.ParseMetricSlope(src.Slope)
	if err != nil {
		return domain.PendingMetric{}, err
	}
	return dst, nil
}
