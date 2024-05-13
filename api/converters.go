//==============================================================================
// Copyright (C) 2024-present Alces Flight Ltd.
//
// This file is part of Concertim Metric Reporting Daemon.
//
// This program and the accompanying materials are made available under
// the terms of the Eclipse Public License 2.0 which is available at
// <https://www.eclipse.org/legal/epl-2.0>, or alternative license
// terms made available by Alces Flight Ltd - please direct inquiries
// about licensing to licensing@alces-flight.com.
//
// Concertim Metric Reporting Daemon is distributed in the hope that it will be useful, but
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
// IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
// OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
// PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
// details.
//
// You should have received a copy of the Eclipse Public License 2.0
// along with Concertim Metric Reporting Daemon. If not, see:
//
//  https://opensource.org/licenses/EPL-2.0
//
// For more information on Concertim Metric Reporting Daemon, please visit:
// https://github.com/openflighthpc/concertim-metric-reporting-daemon
//==============================================================================

package api

import (
	"time"

	"github.com/openflighthpc/concertim-metric-reporting-daemon/domain"
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
