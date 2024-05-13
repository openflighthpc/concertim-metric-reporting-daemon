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

	"github.com/openflighthpc/concertim-metric-reporting-daemon/domain"
)

type uniqueMetric struct {
	Id     string `json:"id"`
	Max    any    `json:"max"`
	Min    any    `json:"min"`
	Name   string `json:"name"`
	Nature string `json:"nature"`
	Units  string `json:"units"`
}

// getUniqueMetrics returns a JSON list of unique metrics.  The uniqueness of
// metrics is determined by its metric name.  The format of the JSON is as follows:
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity",
//	    "units": "",
//	    "nature": "volatile",
//	    "min": 32,
//	    "max": 64
//	  },
//	  ...
//	]
func (s *Server) getUniqueMetrics(rw http.ResponseWriter, r *http.Request) {
	metrics, err := s.app.CurrentRepo.GetUniqueMetrics()
	if err != nil {
		if errors.Is(err, domain.ErrWaitingOnProcessingRun) {
			ServiceUnavailable(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	body := []uniqueMetric{}
	for _, metric := range metrics {
		um := uniqueMetric{
			Id:     metric.Name,
			Name:   metric.Name,
			Nature: metric.Nature,
			Units:  metric.Units,
			Min:    metric.Min,
			Max:    metric.Max,
		}
		body = append(body, um)
	}
	renderJSON(body, http.StatusOK, rw)
}
