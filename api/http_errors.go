package api

import (
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// ErrorResponse is the response type for any error messages.  It is loosely
// based on the JSON:API error object.
type ErrorResponse struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// InternalError responds with an internal error.
func InternalError(rw http.ResponseWriter, r *http.Request, err error, logMsg string) {
	if logMsg == "" {
		logMsg = http.StatusText(http.StatusInternalServerError)
	}
	respondWithError(rw, r, err, http.StatusInternalServerError, logMsg)
}

// BadRequest responds with a bad request.
func BadRequest(rw http.ResponseWriter, r *http.Request, err error, logMsg string) {
	if logMsg == "" {
		logMsg = http.StatusText(http.StatusBadRequest)
	}
	respondWithError(rw, r, err, http.StatusBadRequest, logMsg)
}

func respondWithError(rw http.ResponseWriter, r *http.Request, err error, status int, logMsg string) {
	logger := hlog.FromRequest(r)
	logger.Info().Err(err).Msg(logMsg)
	resp := ErrorResponse{
		Status: status,
		Title:  logMsg,
		Detail: err.Error(),
	}
	renderJSON(resp, status, rw)
}
