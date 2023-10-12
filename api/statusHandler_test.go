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
