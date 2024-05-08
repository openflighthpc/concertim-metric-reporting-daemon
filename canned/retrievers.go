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
