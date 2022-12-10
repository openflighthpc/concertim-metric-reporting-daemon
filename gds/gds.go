// Package gds provide a Ganglia Data Source Server.  The data it reports is
// current canned.
package gds

import (
	"errors"
	"net"

	"github.com/rs/zerolog"

	"github.com/alces-flight/concertim-mrapi/db"
)

// Server is a wrapper around a net.TCPListener it responds to every
// connection with Ganglia compliant XML.
type Server struct {
	addr      *net.TCPAddr
	logger    zerolog.Logger
	stopChan  chan struct{}
	tcpServer *net.TCPListener
	db        db.DB
}

// New returns a new Server.
func New(logger zerolog.Logger, db db.DB) *Server {
	addr := &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8678,
	}
	return &Server{
		addr:      addr,
		logger:    logger.With().Str("component", "gds").Logger(),
		tcpServer: nil,
		stopChan:  make(chan struct{}),
		db:        db,
	}
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
			output, err := generateOutput(gds.db.GetAll())
			if err != nil {
				gds.logger.Error().Err(err).Msg("Failed to generate output")
			} else {
				if _, err := conn.Write(append(output, []byte("\n")...)); err != nil {
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
