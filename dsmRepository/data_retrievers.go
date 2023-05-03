package dsmRepository

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"

	"github.com/rs/zerolog"
)

// Script retrieves the data source map by running the script
// specified at Path.
type Script struct {
	Args   []string
	Path   string
	Logger zerolog.Logger
}

func (e *Script) getNewData() (map[string]string, error) {
	args := e.Args
	if args == nil {
		args = []string{}
	}
	cmd := exec.Command(e.Path, args...)
	e.Logger.Debug().Str("cmd", cmd.String()).Msg("running")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	parser := Parser{Logger: e.Logger}
	return parser.parseJSON(out)
}

// JSONFileRetreiver retrieves the data source map from a pre-poulated JSON
// file.
type JSONFileRetreiver struct {
	Path   string
	Logger zerolog.Logger
}

func (j *JSONFileRetreiver) getNewData() (map[string]string, error) {
	data, err := ioutil.ReadFile(j.Path)
	if err != nil {
		return nil, err
	}
	parser := Parser{Logger: j.Logger}
	return parser.parseJSON(data)
}

// Parser parses the data provided by a dataRetriever into a
// map[string]string.
//
// The data is expected to be a JSON object with string keys as the device
// name and string values as the device's map to host.
type Parser struct {
	Logger zerolog.Logger
}

func (p *Parser) parseJSON(data []byte) (map[string]string, error) {
	var raw interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	bar := raw.(map[string]interface{})
	newData := map[string]string{}
	for deviceName, mapToHost := range bar {
		mapToHostStr, ok := mapToHost.(string)
		if !ok {
			p.Logger.Warn().Interface("mapToHost", mapToHost).Msg("Could not convert to string")
			continue
		}
		newData[deviceName] = mapToHostStr
	}
	return newData, nil
}
