// Package api provides a HTTP API server.
//
// The server supports adding metrics to hosts.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Server is a wrapper around a net/http.Server.
type Server struct {
	logger     zerolog.Logger
	httpServer *http.Server
	repo       domain.Repository
}

// NewServer returns an *http.Server configured as an API server.
func NewServer(logger zerolog.Logger, repo domain.Repository) *Server {
	return &Server{
		logger: logger.With().Str("component", "api").Logger(),
		repo:   repo,
	}
}

func (s *Server) ListenAndServe() error {
	addr := ":3000"
	server := http.Server{
		Addr:         addr,
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 100,
		Handler:      s.addRoutes(),
	}
	s.httpServer = &server
	s.logger.Info().Str("address", server.Addr).Msg("API server listening")
	return server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) addRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(hlog.NewHandler(s.logger))
	r.Use(logMiddleware())
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(middleware.CleanPath)

	r.Put("/{hostName}/metrics", s.putMetricHandler)

	return r
}

type putMetricRequest struct {
	Name  string `json:"name"`
	Val   any    `json:"value"`
	Units string `json:"units"`
	Type  string `json:"type"`
	Slope string `json:"slope"`
	TTL   uint   `json:"ttl"`
}

type putMetricResponse struct {
	Status int           `json:"status"`
	Metric domain.Metric `json:"metric"`
}

func (s *Server) putMetricHandler(rw http.ResponseWriter, r *http.Request) {
	putMetric := &putMetricRequest{}
	err := parseJSONBody(putMetric, rw, r)
	if err != nil {
		// The correct response has already been sent by parseJSONBody.
		return
	}
	metric, err := DomainMetricFromPutMetric(*putMetric)
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	err = domain.AddMetric(s.repo, metric, chi.URLParam(r, "hostName"))
	if err != nil {
		logger := hlog.FromRequest(r)
		logger.Debug().Err(err).Msg("adding metric")
		renderJSON("", http.StatusInternalServerError, rw)
		return
	}

	body := putMetricResponse{Status: http.StatusOK, Metric: metric}
	renderJSON(body, http.StatusOK, rw)
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
