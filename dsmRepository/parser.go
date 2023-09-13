package dsmRepository

import (
	"encoding/json"
	"fmt"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Parser parses the data provided by a dataRetriever into a
// map[domain.HostId]domain.DSM and a map[domain.DSM]domain.HostId.
type Parser struct {
	Logger zerolog.Logger
}

func (p *Parser) parseJSON(data []byte) (map[domain.HostId]domain.DSM, map[domain.DSM]domain.HostId, error) {
	p.Logger.Debug().Int("bytes", len(data)).Msg("parsing JSON")
	var raw interface{}
	var dsmToHostId  map[domain.DSM]domain.HostId

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, nil, err
	}
	dsmToHostId, err = p.parseHostIdMap(raw)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing data source map")
	}

	hostIdToDSM := make(map[domain.HostId]domain.DSM, len(dsmToHostId))
	for dsm, hostId := range dsmToHostId {
		hostIdToDSM[hostId] = dsm
	}

	return hostIdToDSM, dsmToHostId, nil
}

func (p *Parser) parseHostIdMap(data any) (map[domain.DSM]domain.HostId, error) {
	p.Logger.Debug().Any("data", data).Msg("parsing data source map")
	dsmMap := map[domain.DSM]domain.HostId{}

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
			for hName, hostId := range hostMap {
				hostId, ok := hostId.(string)
				if !ok {
					return nil, fmt.Errorf("unexpected type for hostId")
				}
				dsm := domain.DSM{
					GridName:    gName,
					ClusterName: cName,
					HostName:    hName,
				}
				dsmMap[dsm] = domain.HostId(hostId)
			}
		}
	}
	return dsmMap, nil
}
