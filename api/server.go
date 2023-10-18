// Package api provides a HTTP API server.
//
// The server supports adding metrics to hosts.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Server is a wrapper around a net/http.Server.
type Server struct {
	app        *domain.Application
	config     config.API
	httpServer *http.Server
	logger     zerolog.Logger
	tokenAuth  *jwtauth.JWTAuth
	Router     chi.Router
}

// NewServer returns an *http.Server configured as an API server.
func NewServer(logger zerolog.Logger, app *domain.Application, config config.API) *Server {
	server := Server{
		app:       app,
		config:    config,
		logger:    logger.With().Str("component", "http-api").Logger(),
		tokenAuth: jwtauth.New("HS256", config.JWTSecret, nil),
	}
	server.addRoutes()
	return &server
}

// ListenAndServe runs the HTTP API server.
func (s *Server) ListenAndServe() error {
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.IP, s.config.Port),
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
		Handler:      s.Router,
	}
	s.httpServer = &server
	s.logger.Info().Str("address", server.Addr).Msg("Listening")
	return server.ListenAndServe()
}

// Shutdown stops the HTTP API server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) addRoutes() chi.Router {
	r := chi.NewRouter()
	s.Router = r
	r.Use(hlog.NewHandler(s.logger))
	r.Use(logMiddleware())
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(middleware.CleanPath)

	r.Get("/status", s.statusHandler)
	r.Group(func(r chi.Router) {
		// Currently, as long as the JWT token can be verified, we allow all
		// access.  Later we probably want to check the claims that are being
		// made.
		r.Use(jwtauth.Verifier(s.tokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Put("/{deviceId}/metrics", s.putMetricHandler)
	})

	// Route to get metrics for a single device.
	r.Get("/devices/{deviceId}/metrics/current", s.getCurrentHostMetrics)
	r.Get("/devices/{deviceId}/metrics/historic", s.getHistoricHostMetricNames)
	r.Get("/devices/{deviceId}/metrics/{metricName}/historic/last/{duration}", s.getHistoricHostMetricValuesLastX)
	r.Get("/devices/{deviceId}/metrics/{metricName}/historic/{startTime}/{endTime}", s.getHistoricHostMetricValues)

	// Routes to get metrics for all devices.
	r.Get("/metrics/unique", s.deprecated(s.getUniqueMetrics))
	r.Get("/metrics/current", s.getUniqueMetrics)
	r.Get("/metrics/historic", s.getHistoricMetricNames)
	r.Get("/metrics/{metricName}/historic/last/{duration}", s.getHistoricMetricValuesLastX)
	r.Get("/metrics/{metricName}/historic/{startTime}/{endTime}", s.getHistoricMetricValues)
	r.Get("/metrics/{metricName}/current", s.getMetricValues)
	r.Get("/metrics/{metricName}/values", s.deprecated(s.getMetricValues))

	return r
}

type putMetricRequest struct {
	Name  string `json:"name"  validate:"required,notblank"`
	Val   any    `json:"value" validate:"required"`
	Units string `json:"units" validate:"excludesall=<>'\"&"`
	Type  string `json:"type"  validate:"required,oneof=string int8 uint8 int16 uint16 int32 uint32 float double"`
	Slope string `json:"slope" validate:"required,oneof=zero positive negative both derivative"`
	TTL   int    `json:"ttl"   validate:"required,min=1"`
}

type putMetricResponse struct {
	Status int `json:"status"`
}

func (s *Server) putMetricHandler(rw http.ResponseWriter, r *http.Request) {
	putMetric := &putMetricRequest{}
	err := parseJSONBody(putMetric, rw, r)
	if err != nil {
		// The correct response has already been sent by parseJSONBody.
		return
	}
	metric, err := domainMetricFromPutMetric(*putMetric, s.logger)
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	err = s.app.AddMetric(metric, domain.HostId(chi.URLParam(r, "deviceId")))
	if errors.Is(err, domain.ErrUnknownHost) {
		body := ErrorsPayload{
			Status: http.StatusNotFound,
			Errors: []*ErrorObject{{Title: "Host Not Found", Detail: err.Error()}},
		}
		renderJSON(body, http.StatusNotFound, rw)
		return
	} else if err != nil {
		logger := hlog.FromRequest(r)
		logger.Debug().Err(err).Send()
		renderJSON("", http.StatusInternalServerError, rw)
		return
	}

	body := putMetricResponse{Status: http.StatusOK}
	renderJSON(body, http.StatusOK, rw)
}

func (s *Server) statusHandler(rw http.ResponseWriter, r *http.Request) {
	type response struct {
		Status int `json:"status"`
	}
	body := response{Status: http.StatusOK}
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
