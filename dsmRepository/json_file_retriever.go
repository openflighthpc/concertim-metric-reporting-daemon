package dsmRepository

import (
	"encoding/json"
	"io/ioutil"

	"github.com/rs/zerolog"
)

// JSONFileRetreiver retrieves the data source map from a pre-poulated JSON
// file.
type JSONFileRetreiver struct {
	Path   string
	Logger zerolog.Logger
}

func (j *JSONFileRetreiver) getNewData() (map[string]string, error) {
	var raw interface{}
	data, err := ioutil.ReadFile(j.Path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}
	bar := raw.(map[string]interface{})
	newData := map[string]string{}
	for deviceName, mapToHost := range bar {
		mapToHostStr, ok := mapToHost.(string)
		if !ok {
			j.Logger.Warn().Interface("mapToHost", mapToHost).Msg("Could not convert to string")
			continue
		}
		newData[deviceName] = mapToHostStr
	}
	return newData, nil
}
