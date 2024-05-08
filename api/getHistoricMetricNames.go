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

import "net/http"

// getHistoricMetricNames returns a JSON list of historic metric names. The
// format of the JSON is as follows:
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity"
//	  },
//	  ...
//	]
func (s *Server) getHistoricMetricNames(rw http.ResponseWriter, r *http.Request) {
	type historicMetricName struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	metricNames, err := s.app.HistoricRepo.ListMetricNames()
	if err != nil {
		InternalError(rw, r, err)
		return
	}
	body := []historicMetricName{}
	for _, metricName := range metricNames {
		um := historicMetricName{
			Id:   metricName,
			Name: metricName,
		}
		body = append(body, um)
	}
	renderJSON(body, http.StatusOK, rw)
}
