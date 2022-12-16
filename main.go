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

	"github.com/alces-flight/concertim-metric-reporting-daemon/api"
	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/dsmRepository"
	"github.com/alces-flight/concertim-metric-reporting-daemon/gds"
	"github.com/alces-flight/concertim-metric-reporting-daemon/repository/memory"
)

func init() {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	isatty := err == nil
	if isatty {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
}

func main() {
	config, err := config.FromFile(config.DefaultPaths)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}
	repository := memory.New(log.Logger)
	dsmRepo := dsmRepository.New(log.Logger, config.DSM)
	app := domain.NewApp(*config, repository, dsmRepo)
	apiServer := api.NewServer(log.Logger, app, config.API)
	gdsServer, err := gds.New(log.Logger, app, config.GDS)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create gds.Server")
	}
	go func() {
		err := apiServer.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Info().Msg("api.Server closed. Waiting for active connections to finish")
		} else if err != nil {
			log.Fatal().Err(err).Msg("api.Server.ListenAndServe")
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
