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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

type metricValue struct {
	Id    string `json:"id"`
	Value any    `json:"value"`
}

// getMetricValues returns a JSON list of current values for metric.
//
//	[
//	  {
//	    "id": "42",
//	    "value": 32
//	  },
//	  ...
//	]
func (s *Server) getMetricValues(rw http.ResponseWriter, r *http.Request) {
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	hosts, err := s.app.CurrentRepo.HostsWithMetric(metricName)
	if err != nil {
		if errors.Is(err, domain.ErrWaitingOnProcessingRun) {
			ServiceUnavailable(rw, r, err)
		} else if errors.Is(err, domain.ErrMetricNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	body := []metricValue{}
	for _, host := range hosts {
		metric, ok := host.Metrics[metricName]
		if !ok {
			s.logger.Warn().
				Stringer("host", host.Id).
				Str("metric", string(metricName)).
				Msg("metric not found for host")
			continue
		}
		castValue, err := castMetricValue(metric)
		if err != nil {
			s.logger.Debug().
				Err(err).
				Str("type", metric.Datatype).
				Str("value", metric.Value).
				Msg("error casting metric value")
			castValue = metric.Value
		}
		mv := metricValue{
			Id:    host.Id.String(),
			Value: castValue,
		}
		body = append(body, mv)
	}
	renderJSON(body, http.StatusOK, rw)
}

func castMetricValue(metric domain.CurrentMetric) (any, error) {
	switch metric.Datatype {
	case "int8", "int16", "int32":
		i, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case "uint8", "uint16", "uint32":
		i, err := strconv.ParseUint(metric.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case "string":
		return metric.Value, nil
	case "float", "double":
		i, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	default:
		return nil, fmt.Errorf("unexpected metric type %s", metric.Datatype)
	}
}
