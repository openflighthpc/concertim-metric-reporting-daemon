package domain

import "time"

// AddMetric adds the given metric for the specified host to the repository.
// If the host has not previously been added it will also be added if its data
// source map to host can be found in the DataSourceMapRepository.
func AddMetric(repo Repository, dsmRepo DataSourceMapRepository, metric Metric, hostName string) error {
	host, ok := repo.GetHost(hostName)
	if !ok {
		var err error
		host, err = addHost(repo, dsmRepo, hostName)
		if err != nil {
			return err
		}
	}
	err := repo.PutMetric(host, metric)
	return err
}

func addHost(repo Repository, dsmRepo DataSourceMapRepository, hostName string) (Host, error) {
	mapToHost, ok := dsmRepo.Get(hostName)
	if !ok {
		return Host{}, UnknownHost{HostName: hostName}
	}
	host := Host{
		DeviceName: hostName,
		DSMName:    mapToHost,
		Reported:   time.Now(),
		DMax:       60,
	}
	err := repo.PutHost(host)
	if err != nil {
		return Host{}, err
	}
	return host, nil
}
