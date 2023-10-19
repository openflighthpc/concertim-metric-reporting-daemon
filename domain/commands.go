package domain

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// AddMetric adds the given metric for the specified host to the repository.
// If the host has not previously been added it will also be added if its data
// source map to host can be found in the DataSourceMapRepository.
func (app *Application) AddMetric(metric ReportedMetric, hostId HostId) error {
	host, ok := app.Repo.GetHost(hostId)
	if !ok {
		var err error
		host, err = app.addHost(hostId)
		if err != nil {
			return errors.Wrap(err, "adding host")
		}
	}
	host.Reported = time.Now()
	err := app.Repo.PutHost(host)
	if err != nil {
		return errors.Wrap(err, "updating host")
	}
	err = app.Repo.PutMetric(host, metric)
	return errors.Wrap(err, "putting metric")
}

// addHost creates a new Host and adds it to the Repository.
//
// The host is only added if a data source map can be found in the
// DataSourceMapRepository.  Otherwise an error is returned.
func (app *Application) addHost(hostId HostId) (ReportedHost, error) {
	dsm, ok := app.dsmRepo.GetDSM(hostId)
	if !ok {
		app.dsmUpdater.UpdateNow()
		dsm, ok = app.dsmRepo.GetDSM(hostId)
	}
	if !ok {
		return ReportedHost{}, fmt.Errorf("%w: %s", ErrUnknownHost, hostId)
	}
	host := ReportedHost{
		Id:       hostId,
		DSM:      dsm,
		Reported: time.Now(),
		DMax:     time.Duration(app.config.HostTTL) * time.Second,
	}
	err := app.Repo.PutHost(host)
	if err != nil {
		return ReportedHost{}, err
	}
	return host, nil
}
