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
}

// NewServer returns an *http.Server configured as an API server.
func NewServer(logger zerolog.Logger, app *domain.Application, config config.API) *Server {
	return &Server{
		app:       app,
		config:    config,
		logger:    logger.With().Str("component", "api").Logger(),
		tokenAuth: jwtauth.New("HS256", []byte(config.JWTSecret), nil),
	}
}

// ListenAndServe runs the HTTP API server.
func (s *Server) ListenAndServe() error {
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.IP, s.config.Port),
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 100,
		Handler:      s.addRoutes(),
	}
	s.httpServer = &server
	s.logger.Info().Str("address", server.Addr).Msg("API server listening")
	return server.ListenAndServe()
}

// Shutdown stops the HTTP API server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) addRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(hlog.NewHandler(s.logger))
	r.Use(logMiddleware())
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(middleware.CleanPath)

	r.Group(func(r chi.Router) {
		// Currently, as long as the JWT token can be verified, we allow all
		// access.  Later we probably want to check the claims that are being
		// made.
		r.Use(jwtauth.Verifier(s.tokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Put("/{deviceName}/metrics", s.putMetricHandler)
	})

	// TODO: Protect the create token route.  We want to allow emma to
	// create tokens.  Emma has already authenticated the user.  Emma
	// should create its own token using the shared secret between it and
	// ctmrd.  This route should authenticate that token.
	r.Post("/token", s.createToken)

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
	metric, err := domainMetricFromPutMetric(*putMetric)
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	err = s.app.AddMetric(metric, chi.URLParam(r, "deviceName"))
	if errors.Is(err, domain.UnknownHost) {
		body := ErrorResponse{Status: "404", Title: "Host Not Found", Detail: err.Error()}
		renderJSON(body, http.StatusNotFound, rw)
		return
	} else if err != nil {
		logger := hlog.FromRequest(r)
		logger.Debug().Err(err).Msg("adding metric")
		renderJSON("", http.StatusInternalServerError, rw)
		return
	}

	body := putMetricResponse{Status: http.StatusOK, Metric: metric}
	renderJSON(body, http.StatusOK, rw)
}

type createTokenRequest struct {
	ExpiresIn time.Duration `json:"expires_in,omitempty"`
}

type createTokenResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

func (s *Server) createToken(rw http.ResponseWriter, r *http.Request) {
	tokenRequest := &createTokenRequest{}
	err := parseJSONBody(tokenRequest, rw, r)
	if err != nil {
		// The correct response has already been sent by parseJSONBody.
		return
	}
	if tokenRequest.ExpiresIn == 0 {
		tokenRequest.ExpiresIn = time.Hour * 24
	}

	claims := map[string]interface{}{}
	jwtauth.SetExpiryIn(claims, tokenRequest.ExpiresIn)
	s.logger.Info().Int64("exp", claims["exp"].(int64)).Send()

	_, tokenString, err := s.tokenAuth.Encode(claims)
	if err != nil {
		logger := hlog.FromRequest(r)
		logger.Debug().Err(err).Msg("Creating token")
		renderJSON("", http.StatusInternalServerError, rw)
		return
	}
	logger := hlog.FromRequest(r)
	logger.Info().Interface("claims", claims).Msg("Created token")

	body := createTokenResponse{Status: http.StatusOK, Token: tokenString}
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
