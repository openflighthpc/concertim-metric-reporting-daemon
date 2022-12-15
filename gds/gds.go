// Package gds provide a Ganglia Data Source Server.
package gds

import (
	"errors"
	"fmt"
	"net"

	"github.com/rs/zerolog"

	"github.com/alces-flight/concertim-mrapi/config"
	"github.com/alces-flight/concertim-mrapi/domain"
)

// Server is a wrapper around a net.TCPListener it responds to every
// connection with Ganglia compliant XML.
type Server struct {
	addr      *net.TCPAddr
	generator *outputGenerator
	logger    zerolog.Logger
	repo      domain.Repository
	stopChan  chan struct{}
	tcpServer *net.TCPListener
}

// New returns a new Server.
func New(logger zerolog.Logger, repo domain.Repository, config config.GDS) (*Server, error) {
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
		logger.Error().Err(err).Msg("Unable to create output generator")
		return nil, err
	}
	return &Server{
		addr:      addr,
		generator: generator,
		logger:    logger.With().Str("component", "gds").Logger(),
		repo:      repo,
		stopChan:  make(chan struct{}),
		tcpServer: nil,
	}, nil
}

// ListenAndServe listens on the configured address and responds to each
// connection with Ganglia XML.
func (gds *Server) ListenAndServe() error {
	listener, err := net.ListenTCP("tcp", gds.addr)
	if err != nil {
		return err
	}
	gds.tcpServer = listener
	gds.logger.Info().Stringer("address", gds.addr).Msg("Listenting")
	go func() {
		<-gds.stopChan
		if err := gds.tcpServer.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			gds.logger.Warn().Err(err).Msg("Error closing TCP server")
		}
	}()
	for {
		conn, err := gds.tcpServer.Accept()
		if err != nil {
			return err
		}
		gds.logger.Info().Stringer("from", conn.RemoteAddr()).Msg("Accepted connection")
		go func() {
			output, err := gds.generator.generate(gds.repo.GetAll())
			if err != nil {
				gds.logger.Error().Err(err).Msg("Failed to generate output")
			} else {
				if _, err := conn.Write(output); err != nil {
					gds.logger.Warn().Err(err).Msg("Sending response")
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
