package gds

import (
	"errors"
	"net"

	"github.com/rs/zerolog"
)

type Server struct {
	addr      *net.TCPAddr
	logger    zerolog.Logger
	stopChan  chan struct{}
	tcpServer *net.TCPListener
}

func (gds *Server) Close() error {
	select {
	case <-gds.stopChan:
		return errors.New("gds.Server already closed")
	default:
		gds.stopChan <- struct{}{}
		return nil
	}
}

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
				c.Write([]byte("OK\n"))
				c.Close()
			}(conn)
		case <-gds.stopChan:
			gds.tcpServer.Close()
			close(queue)
			return nil
		}
	}
}
