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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func Test_Status(t *testing.T) {
	// Setup
	server := NewServer(log.Logger, nil, testAPIConfig)
	req, err := http.NewRequest("GET", "/status", nil)
	assert.NoError(t, err, "unexpected failure building http request")
	rr := httptest.NewRecorder()

	// Action
	server.Router.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code, "expected status code 200")
	assertContentType(t, rr, "application/json")
	assert.JSONEq(t, `{"status": 200}`, rr.Body.String(), "unexpected body")
}

func Test_Status_2(t *testing.T) {
	server := NewServer(log.Logger, nil, testAPIConfig)
	assert.HTTPSuccess(t, server.Router.ServeHTTP, "GET", "/status", nil)
	body := assert.HTTPBody(server.Router.ServeHTTP, "GET", "/status", nil)
	assert.JSONEq(t, `{"status": 200}`, body, "unexpected body")
}
