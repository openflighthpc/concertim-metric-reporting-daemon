//==============================================================================
// Copyright (C) 2024-present Alces Flight Ltd.
//
// This file is part of Concertim Metric Reporting Daemon.
//
// This program and the accompanying materials are made available under
// the terms of the Eclipse Public License 2.0 which is available at
// <https://www.eclipse.org/legal/epl-2.0>, or alternative license
// terms made available by Alces Flight Ltd - please direct inquiries
// about licensing to licensing@alces-flight.com.
//
// Concertim Metric Reporting Daemon is distributed in the hope that it will be useful, but
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
// IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
// OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
// PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
// details.
//
// You should have received a copy of the Eclipse Public License 2.0
// along with Concertim Metric Reporting Daemon. If not, see:
//
//  https://opensource.org/licenses/EPL-2.0
//
// For more information on Concertim Metric Reporting Daemon, please visit:
// https://github.com/openflighthpc/concertim-metric-reporting-daemon
//==============================================================================

package visualizer

import (
	"encoding/json"
	"fmt"

	"github.com/openflighthpc/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Parser parses the data provided by a dataRetriever into a
// map[domain.HostId]domain.DSM and a map[domain.DSM]domain.HostId.
type Parser struct {
	Logger zerolog.Logger
}

func (p *Parser) ParseDSM(data []byte) (map[domain.HostId]domain.DSM, map[domain.DSM]domain.HostId, error) {
	p.Logger.Debug().Int("bytes", len(data)).Msg("parsing JSON")
	var raw interface{}
	var dsmToHostId map[domain.DSM]domain.HostId

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
