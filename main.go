package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"

	"github.com/alces-flight/concertim-mrapi/gds"
)

func newAPIServer(idleConnsClosed <-chan struct{}) *http.Server {
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
	idleConnsClosed := make(chan struct{})
	apiServer := newAPIServer(idleConnsClosed)
	gdsServer := gds.New(log.Logger)
	go func() {
		log.Info().Str("address", apiServer.Addr).Msg("API server listening")
		err := apiServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http.Server.ListenAndServe")
		}
		<-idleConnsClosed
	}()
	go func() {
		err := gdsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("gds.Server.ListenAndServe")
		}
	}()
	gracefulExitSigs := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, gracefulExitSigs...)
	<-sigint
	log.Info().Msg("Closing connections")
	signal.Reset(gracefulExitSigs...)

	if err := apiServer.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("http.Server.Shutdown")
	}
	if err := gdsServer.Close(); err != nil {
		log.Error().Err(err).Msg("gds.Server.Close")
	}
	close(idleConnsClosed)
}
