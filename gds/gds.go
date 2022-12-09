// Package gds provide a Ganglia Data Source Server.  The data it reports is
// current canned.
package gds

import (
	"encoding/xml"
	"errors"
	"net"

	"github.com/rs/zerolog"
)

// Server is a wrapper around a net.TCPListener it responds to every
// connection with Ganglia compliant XML.
type Server struct {
	addr      *net.TCPAddr
	logger    zerolog.Logger
	stopChan  chan struct{}
	tcpServer *net.TCPListener
}

// New returns a new Server.
func New(logger zerolog.Logger) *Server {
	addr := &net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8678,
	}
	return &Server{
		addr:      addr,
		logger:    logger.With().Str("component", "gds").Logger(),
		tcpServer: nil,
		stopChan:  make(chan struct{}),
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
	queue := make(chan net.Conn)
	go func() {
		for {
			conn, err := gds.tcpServer.Accept()
			if err != nil {
				gds.logger.Warn().Err(err).Msg("Accept")
				// I assume here that if we've got an error it is because the
				// TCPServer has been closed.
				break
			} else {
				gds.logger.Info().Stringer("from", conn.RemoteAddr()).Msg("Accepted connection")
				queue <- conn
			}
		}
	}()
	for {
		select {
		case conn := <-queue:
			go func(c net.Conn) {
				if output, err := xml.MarshalIndent(fakeCluster(), "", "  "); err != nil {
					gds.logger.Error().Err(err).Msg("xml.Marshal")
				} else {
					if _, err := c.Write(append(output, []byte("\n")...)); err != nil {
						gds.logger.Warn().Err(err).Msg("Sending response")
					}
				}
				if err := c.Close(); err != nil {
					gds.logger.Warn().Err(err).Msg("Error closing connection")
				}
			}(conn)
		case <-gds.stopChan:
			if err := gds.tcpServer.Close(); err != nil {
				gds.logger.Warn().Err(err).Msg("Error closing TCP server")
			}
			close(queue)
			close(gds.stopChan)
			return nil
		}
	}
}

// Close the server.  Any active connections are dropped.
func (gds *Server) Close() error {
	select {
	case <-gds.stopChan:
		return errors.New("gds.Server already closed")
	default:
		gds.stopChan <- struct{}{}
		return nil
	}
}
