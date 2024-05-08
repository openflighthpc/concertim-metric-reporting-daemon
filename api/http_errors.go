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
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// ErrorsPayload is the response type for any error messages.  It is loosely
// based on the JSON:API error payload.
type ErrorsPayload struct {
	Errors []*ErrorObject `json:"errors"`
	Status int            `json:"status"`
}

// ErrorObject represents a single error.  It is loosely based on the JSON:API
// error object.
type ErrorObject struct {
	Status int    `json:"status,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Source string `json:"source,omitempty"`
}

func (e *ErrorObject) Error() string {
	return fmt.Sprintf("Error: %s %s\n", e.Title, e.Detail)
}

// InternalError responds with an internal error.
func InternalError(rw http.ResponseWriter, r *http.Request, err error) {
	title := http.StatusText(http.StatusInternalServerError)
	logger := hlog.FromRequest(r)
	logger.Info().Err(err).Msg(title)
	respondWithError(rw, r, err, http.StatusInternalServerError, title, title)
}

// BadRequest responds with a bad request.
func BadRequest(rw http.ResponseWriter, r *http.Request, err error, logMsg string) {
	if logMsg != "" {
		logger := hlog.FromRequest(r)
		logger.Info().Err(err).Msg(logMsg)
	}
	title := http.StatusText(http.StatusBadRequest)
	respondWithError(rw, r, err, http.StatusBadRequest, title, logMsg)
}

func ServiceUnavailable(rw http.ResponseWriter, r *http.Request, err error) {
	title := http.StatusText(http.StatusServiceUnavailable)
	respondWithError(rw, r, err, http.StatusServiceUnavailable, title, "")
}

func NotFound(rw http.ResponseWriter, r *http.Request, err error) {
	title := http.StatusText(http.StatusNotFound)
	respondWithError(rw, r, err, http.StatusNotFound, title, "")
}

func respondWithError(rw http.ResponseWriter, r *http.Request, err error, status int, title, logMsg string) {
	resp := ErrorsPayload{
		Errors: []*ErrorObject{{Title: title, Detail: err.Error()}},
		Status: status,
	}
	renderJSON(resp, status, rw)
}
