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
	"net/http"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

// getCurrentHostMetrics returns a JSON list of all current metrics for the
// given host.
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity",
//	    "units": "cups",
//	    "nature": "volatile",
//	    "value": 3,
//	  },
//	  ...
//	]
func (s *Server) getCurrentHostMetrics(rw http.ResponseWriter, r *http.Request) {
	type metricResponse struct {
		Id     string `json:"id"`
		Name   string `json:"name"`
		Nature string `json:"nature"`
		Units  string `json:"units"`
		Value  any    `json:"value"`
	}
	body := []metricResponse{}
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
	metrics, err := s.app.CurrentRepo.GetMetricsForHost(hostId)
	if err != nil {
		if errors.Is(err, domain.ErrWaitingOnProcessingRun) {
			ServiceUnavailable(rw, r, err)
		} else if errors.Is(err, domain.ErrHostNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	for _, metric := range metrics {
		castValue, err := castMetricValue(*metric)
		if err != nil {
			s.logger.Debug().Err(err).Str("type", metric.Datatype).Str("value", metric.Value).Msg("error casting metric value")
			castValue = metric.Value
		}
		mr := metricResponse{
			Id:     metric.Name,
			Name:   metric.Name,
			Nature: metric.Nature,
			Units:  metric.Units,
			Value:  castValue,
		}
		body = append(body, mr)
	}
	renderJSON(body, http.StatusOK, rw)
}
