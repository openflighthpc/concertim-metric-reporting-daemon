// Package memory provides an in-memory repository.  All data it stores will
// be lost when the process exits.
package memory

import (
	"strings"
	"sync"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/rs/zerolog"
)

// Repo is an in-memory repository.  All data it stores will be lost
// when the process exits.
type Repo struct {
	hosts   map[string]HostModel
	metrics map[string]map[string]MetricModel
	mux     sync.Mutex
	logger  zerolog.Logger
}

// PutHost implements the Repository interface.
func (mr *Repo) PutHost(host domain.Host) error {
	mr.logger.Debug().Str("host", host.DeviceName).Msg("Putting host")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.hosts[host.DeviceName] = conv.ModelFromDomainHost(host)
	return nil
}

func (mr *Repo) isHostStored(host domain.Host) bool {
	_, ok := mr.hosts[host.DeviceName]
	return ok
}

// PutMetric implements the Repository interface.
func (mr *Repo) PutMetric(host domain.Host, metric domain.Metric) error {
	mr.logger.Debug().Str("host", host.DeviceName).Str("metric", metric.Name).Msg("Putting metric")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	if !mr.isHostStored(host) {
		return domain.UnknownHost{HostName: host.DeviceName}
	}
	metrics, ok := mr.metrics[host.DeviceName]
	if !ok {
		metrics = make(map[string]MetricModel, 0)
		mr.metrics[host.DeviceName] = metrics
	}
	metrics[metric.Name] = conv.ModelFromDomainMetric(metric)
	return nil
}

// GetAll implements the Repository interface.
func (mr *Repo) GetAll() domain.Cluster {
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.logger.Debug().Msg("Getting all data")
	cluster := domain.Cluster{}
	for _, h := range mr.hosts {
		metrics := mr.metrics[h.DeviceName]
		logHostAndMetrics(mr.logger, h, metrics)
		host := domainHostFromModelHostAndMetrics(h, metrics)
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster
}

// GetHost implements the Repository interface.
func (mr *Repo) GetHost(hostName string) (domain.Host, bool) {
	mr.mux.Lock()
	defer mr.mux.Unlock()
	host, ok := mr.hosts[hostName]
	if !ok {
		return domain.Host{}, false
	}
	return conv.DomainFromModelHost(host), true
}

// New returns an empty in-memory repository.
func New(logger zerolog.Logger) *Repo {
	mr := &Repo{
		hosts:   map[string]HostModel{},
		metrics: map[string]map[string]MetricModel{},
		mux:     sync.Mutex{},
		logger:  logger.With().Str("component", "storage").Logger(),
	}
	return mr
}

func logHostAndMetrics(log zerolog.Logger, h HostModel, metrics map[string]MetricModel) {
	log.Debug().
		Str("host", h.DeviceName).
		Int("metric.count", len(metrics)).
		Func(func(e *zerolog.Event) {
			metricNames := []string{}
			for _, mr := range metrics {
				metricNames = append(metricNames, mr.Name)
			}
			e.Str("metric.names", strings.Join(metricNames, ","))
		}).
		Msg("Found host")
}
