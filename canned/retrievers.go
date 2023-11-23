package canned

import (
	"fmt"
	"os"
	"strings"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/visualizer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// DSMRetriever retrieves the data source map from a pre-poulated JSON
// file.
type DSMRetriever struct {
	Path   string
	Logger zerolog.Logger
}

func (j *DSMRetriever) GetDSM() (map[domain.HostId]domain.DSM, map[domain.DSM]domain.HostId, error) {
	j.Logger.Debug().Str("path", j.Path).Msg("retrieving canned DSM json")
	data, err := os.ReadFile(j.Path)
	if err != nil {
		msg := "reading JSON file"
		if !strings.Contains(err.Error(), j.Path) {
			msg = fmt.Sprintf("%s: %s", msg, j.Path)
		}
		return nil, nil, errors.Wrap(err, msg)
	}
	parser := visualizer.Parser{Logger: j.Logger}
	hostIdToDSM, dsmToHostId, err := parser.ParseDSM(data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing DSM")
	}
	return hostIdToDSM, dsmToHostId, nil
}
