package domain

import "time"

func AddMetric(repo Repository, metric Metric, hostName string) error {
	host, ok := repo.GetHost(hostName)
	if !ok {
		host = Host{
			DeviceName: hostName,
			DSMName:    hostName,
			Reported:   time.Now(),
			DMax:       60,
		}
		err := repo.PutHost(host)
		if err != nil {
			return err
		}
	}
	err := repo.PutMetric(host, metric)
	return err
}
