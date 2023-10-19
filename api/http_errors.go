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
