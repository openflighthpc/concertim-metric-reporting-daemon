package dsmRepository

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/alces-flight/concertim-metric-reporting-daemon/visualizer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// JSONFileRetreiver retrieves the data source map from a pre-poulated JSON
// file.
type JSONFileRetreiver struct {
	Path   string
	Logger zerolog.Logger
}

func (j *JSONFileRetreiver) retrieve() ([]byte, error) {
	j.Logger.Debug().Str("path", j.Path).Msg("retrieving json")
	data, err := ioutil.ReadFile(j.Path)
	if err != nil {
		msg := "reading JSON file"
		if !strings.Contains(err.Error(), j.Path) {
			msg = fmt.Sprintf("%s: %s", msg, j.Path)
		}
		return nil, errors.Wrap(err, msg)
	}
	return data, nil
}

func (j *JSONFileRetreiver) describe() string {
	return fmt.Sprintf("file:%s", j.Path)
}

// visualizerAPIRetriever retrieves the data source map from the Concertim
// Visualizer API.
type visualizerAPIRetriever struct {
	client *visualizer.Client
	logger zerolog.Logger
}

func (r *visualizerAPIRetriever) retrieve() ([]byte, error) {
	data, err := r.client.GetDSM()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *visualizerAPIRetriever) describe() string {
	return r.client.Config.DataSourceMapUrl
}
