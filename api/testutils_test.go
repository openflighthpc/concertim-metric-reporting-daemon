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
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/stretchr/testify/assert"
)

var testAPIConfig = config.API{
	IP:           "127.0.0.1",
	JWTSecret:    []byte("secret"),
	Port:         3000,
	ReadTimeout:  50,
	WriteTimeout: 50,
	IdleTimeout:  50,
}

type fakeDSMRepo struct{}

func (fakeDSMRepo) GetDSM(deviceId domain.HostId) (domain.DSM, bool) {
	if deviceId == "NOPE" {
		return domain.DSM{}, false
	}
	dsm := domain.DSM{
		GridName:    "unspecified",
		ClusterName: "unspecified",
		HostName:    fmt.Sprintf("device:%s", deviceId),
	}
	return dsm, true
}

func (fakeDSMRepo) GetHostId(dsm domain.DSM) (domain.HostId, bool) {
	parts := strings.Split(dsm.HostName, ":")
	return domain.HostId(parts[1]), true
}

func (fakeDSMRepo) Update(_ map[domain.HostId]domain.DSM, _ map[domain.DSM]domain.HostId) (_ error) {
	panic("not implemented") // TODO: Implement
}

var testDSMRepo = fakeDSMRepo{}

func assertContentType(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	t.Helper()
	contentType := rr.Result().Header["Content-Type"]
	if assert.NotNil(t, contentType, "content type header not set") {
		if assert.Len(t, contentType, 1, "content type header set multiple times") {
			assert.Contains(t, contentType[0], expected, "unexpected content type")
		}
	}
}
