// Package main runs the HTTP API server and the TCP GDS server.
package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/alces-flight/concertim-mrapi/gds"
	"github.com/alces-flight/concertim-mrapi/repository/memory"
)

func newAPIServer() *http.Server {
	addr := ":3000"
	server := http.Server{
		Addr:         addr,
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 100,
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			if _, err := rw.Write([]byte("OK\n")); err != nil {
				log.Error().Err(err).Msg("http.ResponseWriter.Write")
			}
		}),
	}
	return &server
}

func init() {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	isatty := err == nil
	if isatty {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
}

func main() {
	repository := memory.New(log.Logger)
	addFakeData(repository)
	apiServer := newAPIServer()
	gdsServer := gds.New(log.Logger, repository)
	go func() {
		log.Info().Str("address", apiServer.Addr).Msg("API server listening")
		err := apiServer.ListenAndServe()
		if err != nil && err == http.ErrServerClosed {
			log.Info().Msg("http.Server closed. Waiting for active connections to finish")
		} else if err != nil {
			log.Fatal().Err(err).Msg("http.Server.ListenAndServe")
		}
	}()
	go func() {
		err := gdsServer.ListenAndServe()
		if err != nil && errors.Is(err, net.ErrClosed) {
			log.Info().Msg("gds.Server closed")
		} else if err != nil {
			log.Fatal().Err(err).Msg("gds.Server.ListenAndServe")
		}
	}()
	gracefulExitSigs := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, gracefulExitSigs...)
	<-sigint
	log.Info().Msg("Closing connections")
	signal.Reset(gracefulExitSigs...)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := gdsServer.Close(); err != nil {
		log.Error().Err(err).Msg("gds.Server.Close")
	}
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("http.Server.Shutdown")
	}
}

func addFakeData(m *memory.Repo) {
	comp001 := domain.Host{Name: "comp001", Reported: time.Now().Add(-2 * time.Hour), TMax: 60 * time.Second, DMax: 60 * time.Second}
	comp002 := domain.Host{Name: "comp002", Reported: time.Now().Add(-3 * time.Hour), TMax: 60 * time.Second, DMax: 60 * time.Second}
	err := m.PutHost(comp001)
	if err != nil {
		log.Logger.Warn().Err(err)
	}
	err = m.PutHost(comp002)
	if err != nil {
		log.Logger.Warn().Err(err)
	}
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "foo",
			Val:    "foobar",
			Units:  "foos",
			Slope:  domain.MetricSlopeZero,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeString,
		},
	)
	if err != nil {
		log.Logger.Warn().Err(err)
	}
	// Duplicate foo metric
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "foo",
			Val:    "FOOBAR",
			Units:  "FOOS",
			Slope:  domain.MetricSlopeZero,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeString,
		},
	)
	if err != nil {
		log.Logger.Warn().Err(err)
	}
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "bar",
			Val:    "12",
			Units:  "bars",
			Slope:  domain.MetricSlopeBoth,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeInt32,
		},
	)
	if err != nil {
		log.Logger.Warn().Err(err)
	}
}
