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

// getHistoricHostMetricValues returns a JSON list of historic metric values
// for the given host and metric between the given start and end times.
//
// [
//
//	  {
//	    "timestamp": 1696431225,
//	    "value": 9020
//		 },
//		 ...
//
// ]
func (s *Server) getHistoricHostMetricValues(rw http.ResponseWriter, r *http.Request) {
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	startTime, err := parseTime(chi.URLParam(r, "startTime"))
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	endTime, err := parseTime(chi.URLParam(r, "endTime"))
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	duration := domain.HistoricMetricDurationFromTimes(startTime, endTime)
	s.fetchAndRenderHostMetrics(rw, r, hostId, metricName, duration)
}

// getHistoricHostMetricValuesLastX returns a JSON list of historic metric
// values for the given host and metric for the last hour/day/quarter.
//
// [
//
//	  {
//	    "timestamp": 1696431225,
//	    "value": 9020
//		 },
//		 ...
//
// ]
func (s *Server) getHistoricHostMetricValuesLastX(rw http.ResponseWriter, r *http.Request) {
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	lastX := chi.URLParam(r, "duration")
	duration, err := domain.HistoricMetricDurationFromString(lastX)
	if err != nil {
		if errors.Is(err, domain.ErrLastXLookupMissingEntry) {
			InternalError(rw, r, err)
		} else {
			BadRequest(rw, r, err, "")
		}
		return
	}
	s.fetchAndRenderHostMetrics(rw, r, hostId, metricName, duration)
}

func (s *Server) fetchAndRenderHostMetrics(
	rw http.ResponseWriter,
	r *http.Request,
	hostId domain.HostId,
	metricName domain.MetricName,
	duration domain.HistoricMetricDuration,
) {
	host, err := s.app.HistoricRepo.GetValuesForHostAndMetric(hostId, metricName, duration)
	if err != nil {
		if errors.Is(err, domain.ErrHostNotFound) {
			NotFound(rw, r, err)
		} else if errors.Is(err, domain.ErrMetricNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	body := []historicValueResponse{}
	metrics, ok := host.Metrics[metricName]
	if !ok {
		NotFound(rw, r, domain.ErrMetricNotFound)
	}
	for _, metric := range metrics {
		body = append(body, historicValueResponseFromHistoricMetric(metric))
	}
	renderJSON(body, http.StatusOK, rw)
}
