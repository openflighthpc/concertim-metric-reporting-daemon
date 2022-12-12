// Package api provides a HTTP API server.
//
// The server supports adding metrics to hosts.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// NewServer returns an *http.Server configured as an API server.
func NewServer(logger zerolog.Logger) *http.Server {
	addr := ":3000"
	server := http.Server{
		Addr:         addr,
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 100,
		Handler:      newRouter(logger),
	}
	return &server
}

func newRouter(logger zerolog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(hlog.NewHandler(logger))
	r.Use(logMiddleware())
	r.Use(hlog.RemoteAddrHandler("ip"))

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		if _, err := rw.Write([]byte("OK\n")); err != nil {
			log.Error().Err(err).Msg("http.ResponseWriter.Write")
		}
	})

	r.Put("/metrics/", putMetricHandler)

	return r
}

func parseBody(params any, rw http.ResponseWriter, r *http.Request) error {
	rawRequestBody, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write([]byte(fmt.Sprintf("error reading body: %s\n", err.Error())))
		return err
	}

	err = json.Unmarshal(rawRequestBody, params)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write([]byte(fmt.Sprintf("error decoding body: %s\n", err.Error())))
		return err
	}

	return nil
}

type putMetricRequest struct {
	Name  string `json:"name"`
	Val   any    `json:"value"`
	Units string `json:"units"`
	Mtype string `json:"type"`
	Slope string `json:"slope"`
	Ttl   uint   `json:"ttl"`
}

func putMetricHandler(rw http.ResponseWriter, r *http.Request) {
	params := &putMetricRequest{}
	err := parseBody(params, rw, r)
	if err != nil {
		logger := hlog.FromRequest(r)
		logger.Warn().Err(err).Msg("parsing body")

		return
	}

	logger := hlog.FromRequest(r)
	logger.Printf("params: %#v", params)

	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "application/json")
	body := struct {
		Status string
		Req    putMetricRequest
	}{
		Status: http.StatusText(http.StatusCreated),
		Req:    *params,
	}

	serializedBody, _ := json.Marshal(body)
	_, _ = rw.Write(append(serializedBody, []byte("\n")...))
}

func logMiddleware() func(http.Handler) http.Handler {
	return hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	})
}
