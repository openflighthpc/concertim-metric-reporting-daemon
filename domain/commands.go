package domain

import (
	"fmt"
	"time"
)

// AddMetric adds the given metric for the specified host to the repository.
// If the host has not previously been added it will also be added if its data
// source map to host can be found in the DataSourceMapRepository.
func (app *Application) AddMetric(metric Metric, hostName Hostname) error {
	host, ok := app.Repo.GetHost(hostName)
	if !ok {
		var err error
		host, err = app.addHost(hostName)
		if err != nil {
			return err
		}
	}
	host.Reported = time.Now()
	err := app.Repo.PutHost(host)
	if err != nil {
		return err
	}
	err = app.Repo.PutMetric(host, metric)
	return err
}

// addHost creates a new Host and adds it to the Repository.
//
// The host is only added if a data source map can be found in the
// DataSourceMapRepository.  Otherwise an error is returned.
func (app *Application) addHost(hostName Hostname) (Host, error) {
	dsm, ok := app.dsmRepo.GetDSM(hostName)
	if !ok {
		return Host{}, fmt.Errorf("%w: %s", UnknownHost, hostName)
	}
	host := Host{
		Name:     hostName,
		DSM:      dsm,
		Reported: time.Now(),
		DMax:     time.Duration(app.config.HostTTL) * time.Second,
	}
	err := app.Repo.PutHost(host)
	if err != nil {
		return Host{}, err
	}
	return host, nil
}
