package api

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/hlog"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func InternalError(rw http.ResponseWriter, r *http.Request, err error, logMsg string) {
	if logMsg == "" {
		logMsg = http.StatusText(http.StatusInternalServerError)
	}
	respondWithError(rw, r, err, http.StatusInternalServerError, logMsg)
}

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
		Status: fmt.Sprintf("%d", status),
		Title:  logMsg,
		Detail: err.Error(),
	}
	renderJSON(resp, status, rw)
}
