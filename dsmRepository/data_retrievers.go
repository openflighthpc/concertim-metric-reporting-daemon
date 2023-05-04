package dsmRepository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Script retrieves the data source map by running the script
// specified at Path.
type Script struct {
	Args   []string
	Path   string
	Logger zerolog.Logger
}

func (e *Script) getNewData() (map[domain.Hostname]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	args := e.Args
	if args == nil {
		args = []string{}
	}
	cmd := exec.Command(e.Path, args...)
	e.Logger.Debug().Str("cmd", cmd.String()).Msg("running")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, err
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

func (j *JSONFileRetreiver) getNewData() (map[domain.Hostname]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	data, err := ioutil.ReadFile(j.Path)
	if err != nil {
		return nil, nil, err
	}
	parser := Parser{Logger: j.Logger}
	return parser.parseJSON(data)
}

// Parser parses the data provided by a dataRetriever into a
// map[domain.Hostname]domain.DSM and a map[domain.DSM]domain.MemcacheKey.
type Parser struct {
	Logger zerolog.Logger
}

func (p *Parser) parseJSON(data []byte) (map[domain.Hostname]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	var raw interface{}

	hostnameMap := map[domain.Hostname]domain.DSM{}
	memcacheMap := map[domain.DSM]domain.MemcacheKey{}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, nil, err
	}
	rawMap := raw.(map[string]interface{})
	for key, nestedMap := range rawMap {
		if key == "hostnameMap" {
			hostnameMap, err = p.parseHostnameMap(nestedMap)
			if err != nil {
				return nil, nil, errors.Wrap(err, "parsing hostnameMap")
			}
		} else if key == "memcacheMap" {
			memcacheMap, err = p.parseMemcacheMap(nestedMap)
			if err != nil {
				return nil, nil, errors.Wrap(err, "parsing memcacheMap")
			}
		} else {
			return nil, nil, fmt.Errorf("unknown key %s", key)
		}
	}

	return hostnameMap, memcacheMap, nil
}

func (p *Parser) parseHostnameMap(data any) (map[domain.Hostname]domain.DSM, error) {
	p.Logger.Debug().Any("data", data).Msg("parsing hostnameMap")
	hostMap := data.(map[string]interface{})
	newData := map[domain.Hostname]domain.DSM{}
	for hostname, mapToHost := range hostMap {
		hName, ok := mapToHost.(string)
		if !ok {
			p.Logger.Warn().Interface("mapToHost", mapToHost).Msg("Could not convert to string")
			continue
		}
		dsm := domain.DSM{
			GridName:    "unspecified",
			ClusterName: "unspecified",
			HostName:    hName,
		}
		newData[domain.Hostname(hostname)] = dsm
	}
	return newData, nil
}

func (p *Parser) parseMemcacheMap(data any) (map[domain.DSM]domain.MemcacheKey, error) {
	p.Logger.Debug().Any("data", data).Msg("parsing memcacheMap")
	dsmMap := map[domain.DSM]domain.MemcacheKey{}

	gridMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for gridMap")
	}

	for gName, clusterMap := range gridMap {
		clusterMap, ok := clusterMap.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unexpected type for clusterMap")
		}
		for cName, hostMap := range clusterMap {
			hostMap, ok := hostMap.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("unexpected type for hostMap")
			}
			for hName, memcacheKey := range hostMap {
				memcacheKey, ok := memcacheKey.(string)
				if !ok {
					return nil, fmt.Errorf("not of expected type")
				}
				dsm := domain.DSM{
					GridName:    gName,
					ClusterName: cName,
					HostName:    hName,
				}
				dsmMap[dsm] = domain.MemcacheKey(memcacheKey)
			}
		}
	}
	return dsmMap, nil
}
