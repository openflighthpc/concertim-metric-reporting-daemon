package dsmRepository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

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

func (e *Script) getNewData() (map[domain.HostId]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	args := e.Args
	if args == nil {
		args = []string{}
	}
	cmd := exec.Command(e.Path, args...)
	e.Logger.Debug().Str("cmd", cmd.String()).Msg("retrieving json")
	out, err := cmd.Output()
	if err != nil {
		msg := "executing script"
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if strings.Contains(exitErr.Error(), e.Path) || strings.Contains(string(exitErr.Stderr), e.Path) {
				return nil, nil, errors.Wrapf(exitErr, "%s: %s", msg, exitErr.Stderr)
			}
			return nil, nil, errors.Wrapf(exitErr, "%s: %s: %s", msg, e.Path, exitErr.Stderr)
		}
		return nil, nil, errors.Wrap(err, msg)
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

func (j *JSONFileRetreiver) getNewData() (map[domain.HostId]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	j.Logger.Debug().Str("path", j.Path).Msg("retrieving json")
	data, err := ioutil.ReadFile(j.Path)
	if err != nil {
		msg := "reading JSON file"
		if !strings.Contains(err.Error(), j.Path) {
			msg = fmt.Sprintf("%s: %s", msg, j.Path)
		}
		return nil, nil, errors.Wrap(err, msg)
	}
	parser := Parser{Logger: j.Logger}
	return parser.parseJSON(data)
}

// Parser parses the data provided by a dataRetriever into a
// map[domain.HostId]domain.DSM and a map[domain.DSM]domain.MemcacheKey.
type Parser struct {
	Logger zerolog.Logger
}

func (p *Parser) parseJSON(data []byte) (map[domain.HostId]domain.DSM, map[domain.DSM]domain.MemcacheKey, error) {
	p.Logger.Debug().Int("bytes", len(data)).Msg("parsing JSON")
	var raw interface{}

	deviceIdToGangliaHostName := map[domain.HostId]domain.DSM{}
	dsmToMemcacheKey := map[domain.DSM]domain.MemcacheKey{}

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, nil, err
	}
	rawMap := raw.(map[string]interface{})
	for key, nestedMap := range rawMap {
		if key == "deviceIdToGangliaHostName" {
			deviceIdToGangliaHostName, err = p.parseDeviceIdToGangliaHostName(nestedMap)
			if err != nil {
				return nil, nil, errors.Wrap(err, "parsing deviceIdToGangliaHostName")
			}
		} else if key == "dsmToMemcacheKey" {
			dsmToMemcacheKey, err = p.parseMemcacheMap(nestedMap)
			if err != nil {
				return nil, nil, errors.Wrap(err, "parsing dsmToMemcacheKey")
			}
		} else {
			return nil, nil, fmt.Errorf("unknown key %s", key)
		}
	}

	return deviceIdToGangliaHostName, dsmToMemcacheKey, nil
}

func (p *Parser) parseDeviceIdToGangliaHostName(data any) (map[domain.HostId]domain.DSM, error) {
	p.Logger.Debug().Any("data", data).Msg("parsing deviceIdToGangliaHostName")
	hostMap := data.(map[string]interface{})
	newData := map[domain.HostId]domain.DSM{}
	for deviceId, gangliaHostName := range hostMap {
		hName, ok := gangliaHostName.(string)
		if !ok {
			p.Logger.Warn().
				Str("deviceId", deviceId).
				Interface("gangliaHostName", gangliaHostName).
				Msg("Could not convert to string")
			continue
		}
		dsm := domain.DSM{
			GridName:    "unspecified",
			ClusterName: "unspecified",
			HostName:    hName,
		}
		newData[domain.HostId(deviceId)] = dsm
	}
	return newData, nil
}

func (p *Parser) parseMemcacheMap(data any) (map[domain.DSM]domain.MemcacheKey, error) {
	p.Logger.Debug().Any("data", data).Msg("parsing dsmToMemcacheKey")
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
					return nil, fmt.Errorf("unexpected type for memcacheKey")
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
