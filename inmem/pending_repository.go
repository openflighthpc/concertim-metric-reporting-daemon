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

package inmem

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

var _ domain.PendingRepository = (*PendingRepository)(nil)

// PendingRepository implements the domain.PendingRepository interface.  It
// stores the reported hosts and their metrics pending their processing.
type PendingRepository struct {
	hosts  map[domain.HostId]domain.PendingHost
	mux    sync.Mutex
	logger zerolog.Logger
}

// PutHost adds a Host to the repository.  If the Host has already been
// added it will be updated.
func (pr *PendingRepository) PutHost(host domain.PendingHost) error {
	pr.logger.Debug().Stringer("host", host.Id).Msg("Putting host")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.hosts[host.Id] = host
	return nil
}

func (pr *PendingRepository) isHostStored(hostId domain.HostId) bool {
	_, ok := pr.hosts[hostId]
	return ok
}

// PutMetric adds a Metric to the repository for a previously added Host.
//
// If the Metric has already been added it will be updated.
//
// If the Host has not been previously added an UnknownHost error is
// returned.
func (pr *PendingRepository) PutMetric(host domain.PendingHost, metric domain.PendingMetric) error {
	pr.logger.Debug().Stringer("host", host.Id).Str("metric", metric.Name).Msg("Putting metric")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	if !pr.isHostStored(host.Id) {
		return fmt.Errorf("%w: %s", domain.ErrUnknownHost, host.Id)
	}
	host.Metrics[domain.MetricName(metric.Name)] = metric
	return nil
}

// GetAll returns a slice of all Hosts added to the repository, populated
// with all of their Metrics.
func (pr *PendingRepository) GetAll() []domain.PendingHost {
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.logger.Debug().Msg("Getting all data")
	hosts := make([]domain.PendingHost, 0, len(pr.hosts))
	for _, host := range pr.hosts {
		logHost(pr.logger, host)
		hosts = append(hosts, host)
	}

	return hosts
}

func (pr *PendingRepository) GetHost(hostId domain.HostId) (domain.PendingHost, bool) {
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.logger.Debug().Stringer("host", hostId).Msg("Getting host")
	return pr.getHost(hostId)
}

func (pr *PendingRepository) getHost(hostId domain.HostId) (domain.PendingHost, bool) {
	host, ok := pr.hosts[hostId]
	if !ok {
		return domain.PendingHost{}, false
	}
	return host, true
}

// UpdateLastProcessed updates the metric's LastProcessed field.
func (pr *PendingRepository) UpdateLastProcessed(hostId domain.HostId, metricName domain.MetricName, t time.Time) error {
	pr.logger.Debug().Stringer("host", hostId).Str("metric", string(metricName)).Time("last processed", t).Msg("updating last processed")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	host, ok := pr.getHost(hostId)
	if !ok {
		return fmt.Errorf("%w: %s", domain.ErrUnknownHost, hostId)
	}
	metric, ok := host.Metrics[metricName]
	if !ok {
		return fmt.Errorf("%w: %s", domain.ErrMetricNotFound, metricName)
	}
	metric.LastProcessed = &t
	host.Metrics[metricName] = metric
	return nil
}

// New returns an empty in-memory pending repository.
func NewPendingRepository(logger zerolog.Logger) *PendingRepository {
	mr := &PendingRepository{
		hosts:  map[domain.HostId]domain.PendingHost{},
		mux:    sync.Mutex{},
		logger: logger.With().Str("component", "pending-repo").Logger(),
	}
	return mr
}

func logHost(log zerolog.Logger, h domain.PendingHost) {
	log.Debug().
		Stringer("host", h.Id).
		Int("metric.count", len(h.Metrics)).
		Func(func(e *zerolog.Event) {
			metricNames := []string{}
			for _, mr := range h.Metrics {
				metricNames = append(metricNames, mr.Name)
			}
			e.Str("metric.names", strings.Join(metricNames, ","))
		}).
		Msg("Found host")
}
