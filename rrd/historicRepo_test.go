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

package rrd

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

type fakeDSMRepo struct{}

func (fakeDSMRepo) GetDSM(deviceId domain.HostId) (domain.DSM, bool) {
	if deviceId == "NOPE" {
		return domain.DSM{}, false
	}
	dsm := domain.DSM{
		GridName:    "unspecified",
		ClusterName: "unspecified",
		HostName:    fmt.Sprintf("device:%s", deviceId),
	}
	return dsm, true
}

func (fakeDSMRepo) GetHostId(dsm domain.DSM) (domain.HostId, bool) {
	parts := strings.Split(dsm.HostName, ":")
	return domain.HostId(parts[1]), true
}

func (fakeDSMRepo) Update(_ map[domain.HostId]domain.DSM, _ map[domain.DSM]domain.HostId) (_ error) {
	panic("not implemented") // TODO: Implement
}

var dsmRepo = fakeDSMRepo{}

func Test_ListMetricNames(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		expected  []string
	}{
		{
			name:      "non-existent directory has no metrics",
			directory: "non-existent",
			expected:  []string{},
		},
		{
			name:      "empty directory has no metrics",
			directory: "empty",
			expected:  []string{},
		},
		{
			name:      "directory with one metric has one metric",
			directory: "power.level-only",
			expected:  []string{"power.level"},
		},
		{
			name:      "directory with multiple metrics returns them all",
			directory: "multiple",
			expected: []string{
				"caffeine.capacity",
				"caffeine.consumption",
				"caffeine.level",
			},
		},
		{
			name:      "non-rrd files are ignored",
			directory: "non-rrds",
			expected:  []string{"power.level"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			rrdDir := fmt.Sprintf("testdata/list-metric-names/%s/", tt.directory)
			config := config.RRD{
				ClusterName: "unspecified",
				GridName:    "unspecified",
				Directory:   rrdDir,
				ToolPath:    "/usr/bin/rrdtool",
			}
			repo := NewHistoricRepo(log.Logger, config, dsmRepo)

			// Action
			output, err := repo.ListMetricNames()

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func Test_ListHostMetricNames(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		hostId    string
		expected  []string
	}{
		{
			name:      "non-existent directory has no metrics",
			directory: "non-existent",
			hostId:    "1",
			expected:  []string{},
		},
		{
			name:      "empty directory has no metrics",
			directory: "empty",
			hostId:    "1",
			expected:  []string{},
		},
		{
			name:      "directory with one metric has one metric",
			directory: "power.level-only",
			hostId:    "1",
			expected:  []string{"power.level"},
		},
		{
			name:      "directory with multiple metrics returns them all",
			directory: "multiple",
			hostId:    "1",
			expected: []string{
				"caffeine.capacity",
				"caffeine.consumption",
				"caffeine.level",
			},
		},
		{
			name:      "non-rrd files are ignored",
			directory: "non-rrds",
			hostId:    "1",
			expected:  []string{"power.level"},
		},
		{
			name:      "only this hosts RRD files are included",
			directory: "multiple-hosts",
			hostId:    "1",
			expected:  []string{"caffeine.level", "power.level"},
		},
		{
			name:      "only this hosts RRD files are included",
			directory: "multiple-hosts",
			hostId:    "2",
			expected:  []string{"caffeine.capacity", "caffeine.consumption"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			rrdDir := fmt.Sprintf("testdata/list-host-metric-names/%s/", tt.directory)
			config := config.RRD{
				ClusterName: "unspecified",
				GridName:    "unspecified",
				Directory:   rrdDir,
				ToolPath:    "/usr/bin/rrdtool",
			}
			repo := NewHistoricRepo(log.Logger, config, dsmRepo)

			// Action
			output, err := repo.ListHostMetricNames(domain.HostId(tt.hostId))

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func Test_GetValuesForHostAndMetric(t *testing.T) {
	tests := []struct {
		name            string
		hostId          string
		metric          string
		startTime       int64
		endTime         int64
		expectedMetrics []*domain.HistoricMetric
		expectedErr     error
	}{
		{
			name:        "non-existent host returns ErrHostNotFound",
			hostId:      "NOPE",
			expectedErr: domain.ErrHostNotFound,
		},
		{
			name:        "non-existent metric returns ErrMetricNotFound",
			hostId:      "1",
			metric:      "NOPE",
			expectedErr: domain.ErrMetricNotFound,
		},
		{
			name:      "returns expected values 1",
			hostId:    "1",
			metric:    "power.level",
			startTime: 1696431210,
			endTime:   1696431300,
			expectedMetrics: []*domain.HistoricMetric{
				{
					Timestamp: 1696431225,
					Value:     9020,
				},
				{
					Timestamp: 1696431240,
					Value:     9020,
				},
				{
					Timestamp: 1696431255,
					Value:     9020,
				},
				{
					Timestamp: 1696431270,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431285,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431300,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431315,
					Value:     math.NaN(),
				},
			},
		},
		{
			name:      "returns expected values 2",
			hostId:    "2",
			metric:    "power.level",
			startTime: 1696431210,
			endTime:   1696431300,
			expectedMetrics: []*domain.HistoricMetric{
				{
					Timestamp: 1696431225,
					Value:     8956,
				},
				{
					Timestamp: 1696431240,
					Value:     8956,
				},
				{
					Timestamp: 1696431255,
					Value:     8956,
				},
				{
					Timestamp: 1696431270,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431285,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431300,
					Value:     math.NaN(),
				},
				{
					Timestamp: 1696431315,
					Value:     math.NaN(),
				},
			},
		},
		{
			name:      "returns expected values 3",
			hostId:    "2",
			metric:    "power.level",
			startTime: 1696431100,
			endTime:   1696431150,
			expectedMetrics: []*domain.HistoricMetric{
				{
					Timestamp: 1696431105,
					Value:     8969,
				},
				{
					Timestamp: 1696431120,
					Value:     8969,
				},
				{
					Timestamp: 1696431135,
					Value:     8967.2666,
				},
				{
					Timestamp: 1696431150,
					Value:     8956,
				},
				{
					Timestamp: 1696431165,
					Value:     8956,
				},
			},
		},
	}

	for _, tt := range tests {
		// tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			rrdDir := "testdata/get-values-for-host-and-metric/"
			config := config.RRD{
				ClusterName: "unspecified",
				GridName:    "unspecified",
				Directory:   rrdDir,
				ToolPath:    "/usr/bin/rrdtool",
			}
			repo := NewHistoricRepo(log.Logger, config, dsmRepo)

			// Action
			historicHost, err := repo.GetValuesForHostAndMetric(
				domain.HostId(tt.hostId),
				domain.MetricName(tt.metric),
				domain.HistoricMetricDuration{
					Start: fmt.Sprint(tt.startTime),
					End:   fmt.Sprint(tt.endTime),
				},
			)

			// Assertions
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)

				// Unfortunately, metrics can contain math.NaN values which
				// makes testing for equality awkward.
				assert.Equal(t, domain.HostId(tt.hostId), historicHost.Id, "historic host does not have the expected Id")
				assert.Equal(t, domain.DSM{GridName: "unspecified", ClusterName: "unspecified", HostName: fmt.Sprintf("device:%s", tt.hostId)}, historicHost.DSM, "historic host does not have the expected DSM")

				assert.Len(t, historicHost.Metrics, 1, "values for only a single metric should be returned")
				for name, metrics := range historicHost.Metrics {
					assert.Equal(t, domain.MetricName(tt.metric), name)
					assert.Len(t, metrics, len(tt.expectedMetrics), "incorrect number of metric values returned")
					for i, em := range tt.expectedMetrics {
						assert.Equal(t, em.Timestamp, metrics[i].Timestamp)
						assert.InDelta(t, em.Value, metrics[i].Value, 0.0001)
					}
				}
			}
		})
	}
}
