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

package domain

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Application is the application container.  It holds references to the
// various singleton components of the system such as the pending, current and
// historic repositories.

// It provides a number of methods that coordinate the interaction between the
// repositories.
type Application struct {
	pendingRepo  PendingRepository
	dsmRepo      DataSourceMapRepository
	dsmUpdater   DataSourceMapRepoUpdater
	CurrentRepo  CurrentRepository
	HistoricRepo HistoricRepository
}

// NewApp returns a newly configured Application.
func NewApp(
	pendingRepo PendingRepository,
	dsmRepo DataSourceMapRepository,
	dsmUpdater DataSourceMapRepoUpdater,
	currentRepo CurrentRepository,
	historicRepo HistoricRepository,
) *Application {
	return &Application{
		pendingRepo:  pendingRepo,
		dsmRepo:      dsmRepo,
		dsmUpdater:   dsmUpdater,
		CurrentRepo:  currentRepo,
		HistoricRepo: historicRepo,
	}
}

// AddPendingMetric adds the given metric for the specified host to the pending
// repository. If the host has not previously been added it will also be added
// if its data source map to host can be found in the DataSourceMapRepository.
func (app *Application) AddPendingMetric(metric PendingMetric, hostId HostId) error {
	host, ok := app.pendingRepo.GetHost(hostId)
	if !ok {
		var err error
		host, err = app.addHost(hostId)
		if err != nil {
			return errors.Wrap(err, "adding host")
		}
	}
	host.Reported = time.Now()
	err := app.pendingRepo.PutHost(host)
	if err != nil {
		return errors.Wrap(err, "updating host")
	}
	err = app.pendingRepo.PutMetric(host, metric)
	return errors.Wrap(err, "putting metric")
}

// addHost creates a new PendingHost and adds it to the pending repository.
//
// The host is only added if a data source map can be found in the
// DataSourceMapRepository.  Otherwise an error is returned.
func (app *Application) addHost(hostId HostId) (PendingHost, error) {
	dsm, ok := app.dsmRepo.GetDSM(hostId)
	if !ok {
		app.dsmUpdater.UpdateNow()
		dsm, ok = app.dsmRepo.GetDSM(hostId)
	}
	if !ok {
		return PendingHost{}, fmt.Errorf("%w: %s", ErrUnknownHost, hostId)
	}
	host := PendingHost{
		Id:       hostId,
		DSM:      dsm,
		Reported: time.Now(),
		Metrics:  map[MetricName]PendingMetric{},
	}
	err := app.pendingRepo.PutHost(host)
	if err != nil {
		return PendingHost{}, err
	}
	return host, nil
}
