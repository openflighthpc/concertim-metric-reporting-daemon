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
	IP:        "127.0.0.1",
	JWTSecret: []byte("secret"),
	Port:      3000,
	Timeout:   50,
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
