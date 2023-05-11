// Package gds provide a Ganglia Data Source Server.
package gds

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
)

// Server is a wrapper around a net.TCPListener it responds to every
// connection with Ganglia compliant XML.
type Server struct {
	// Server sends a value to this chan each time a request has been
	// processed.
	AcceptedChan <-chan struct{}
	acceptedChan chan<- struct{}
	addr         *net.TCPAddr
	app          *domain.Application
	generator    *outputGenerator
	logger       zerolog.Logger
	stopChan     chan struct{}
	tcpServer    *net.TCPListener
}

// New returns a new Server.
func New(logger zerolog.Logger, app *domain.Application, config config.GDS) (*Server, error) {
	ip := net.ParseIP(config.IP)
	if ip == nil {
		return nil, fmt.Errorf("%s is not a valid IP address", config.IP)
	}
	addr := &net.TCPAddr{
		IP:   ip,
		Port: config.Port,
	}
	generator, err := newOutputGenerator(realClock{}, config)
	if err != nil {
		return nil, errors.Wrap(err, "creating gds output generator")
	}
	acceptedChan := make(chan struct{})
	return &Server{
		AcceptedChan: acceptedChan,
		acceptedChan: acceptedChan,
		addr:         addr,
		app:          app,
		generator:    generator,
		logger:       logger.With().Str("component", "ganglia-server").Logger(),
		stopChan:     make(chan struct{}),
		tcpServer:    nil,
	}, nil
}

// ListenAndServe listens on the configured address and responds to each
// connection with Ganglia XML.
func (gds *Server) ListenAndServe() error {
	listener, err := net.ListenTCP("tcp", gds.addr)
	if err != nil {
		return errors.Wrap(err, "starting gds server")
	}
	gds.tcpServer = listener
	gds.logger.Info().Stringer("address", gds.addr).Msg("Listening")
	go func() {
		<-gds.stopChan
		if err := gds.tcpServer.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			gds.logger.Warn().Err(err).Msg("Error closing TCP server")
		}
	}()
	for {
		conn, err := gds.tcpServer.Accept()
		if err != nil {
			return errors.Wrap(err, "gds accepting connection")
		}
		gds.logger.Info().Stringer("from", conn.RemoteAddr()).Msg("Accepted connection")
		go func() {
			output, err := gds.generator.generate(gds.app.Repo.GetAll())
			if err != nil {
				gds.logger.Error().Err(err).Msg("Failed to generate output")
			} else {
				_, err := conn.Write(output)
				if err != nil {
					gds.logger.Warn().Err(err).Msg("Sending response")
				} else {
					gds.acceptedChan <- struct{}{}
				}
			}
			if err := conn.Close(); err != nil {
				gds.logger.Warn().Err(err).Msg("Error closing connection")
			}
		}()
	}
}

// Close the server.  Any active connections are dropped.
func (gds *Server) Close() error {
	select {
	case <-gds.stopChan:
		return errors.New("gds.Server already closed")
	default:
		gds.stopChan <- struct{}{}
		close(gds.stopChan)
		return nil
	}
}
