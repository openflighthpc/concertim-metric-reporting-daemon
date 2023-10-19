package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/rrd"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func Test_getHistoricMetricNames(t *testing.T) {
	// Setup
	rrdConfig := config.RRD{
		ClusterName: "unspecified",
		GridName:    "unspecified",
		Directory:   "testdata/rrds/multiple-hosts-and-metrics/",
		ToolPath:    "/usr/bin/rrdtool",
	}
	historicRepo := rrd.NewHistoricRepo(log.Logger, rrdConfig, testDSMRepo)
	app := domain.Application{
		Repo:         nil,
		ResultRepo:   nil,
		HistoricRepo: historicRepo,
	}
	server := NewServer(log.Logger, &app, testAPIConfig)
	req, err := http.NewRequest("GET", "/metrics/historic", nil)
	assert.NoError(t, err, "unexpected failure building http request")
	rr := httptest.NewRecorder()

	// Action
	server.Router.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code, "expected status code 200")
	assertContentType(t, rr, "application/json")
	expectedJSON := `[
	    { "id": "caffeine.consumption", "name": "caffeine.consumption" },
	    { "id": "caffeine.level", "name": "caffeine.level" },
	    { "id": "power.level", "name": "power.level" }
	]`
	assert.JSONEq(t, expectedJSON, rr.Body.String(), "unexpected body")
}
