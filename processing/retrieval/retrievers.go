package retrieval

import (
	"io"
	"io/ioutil"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// fileRetreiver retrieves the ganglia XML data by reading a file at the given path.
type fileRetreiver struct {
	path   string
	logger zerolog.Logger
}

func (r *fileRetreiver) retrieve() ([]byte, error) {
	r.logger.Debug().Str("path", r.path).Msg("retrieving xml")
	gangliaXML, err := ioutil.ReadFile(r.path)
	if err != nil {
		return nil, errors.Wrap(err, "reading canned xml file")
	}
	return gangliaXML, nil
}

func (r *fileRetreiver) describe() string {
	return r.path
}

// tcpRetriever retrieves the ganglia XML connecting to a Ganglia gmetad
// server.
type tcpRetriever struct {
	addr   *net.TCPAddr
	logger zerolog.Logger
}

func (r *tcpRetriever) retrieve() ([]byte, error) {
	r.logger.Debug().Stringer("addr", r.addr).Msg("retrieving xml")
	conn, err := net.DialTCP("tcp", nil, r.addr)
	if err != nil {
		return nil, errors.Wrap(err, "dialing gmetad")
	}
	reply, err := io.ReadAll(conn)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}
	return reply, nil
}

func (r *tcpRetriever) describe() string {
	return r.addr.String()
}
