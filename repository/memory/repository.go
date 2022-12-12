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
	mr.logger.Debug().Str("host", host.Name).Msg("Putting host")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.hosts[host.Name] = conv.ModelFromDomainHost(host)
	return nil
}

func (mr *Repo) isHostStored(host domain.Host) bool {
	_, ok := mr.hosts[host.Name]
	return ok
}

// PutMetric implements the Repository interface.
func (mr *Repo) PutMetric(host domain.Host, metric domain.Metric) error {
	mr.logger.Debug().Str("host", host.Name).Str("metric", metric.Name).Msg("Putting metric")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	if !mr.isHostStored(host) {
		return domain.UnknownHost{Host: host}
	}
	metrics, ok := mr.metrics[host.Name]
	if !ok {
		metrics = make(map[string]MetricModel, 0)
		mr.metrics[host.Name] = metrics
	}
	metrics[metric.Name] = conv.ModelFromDomainMetric(metric)
	return nil
}

// GetAll implements the Repository interface.
func (mr *Repo) GetAll() domain.Cluster {
	mr.logger.Debug().Msg("Getting all data")
	cluster := domain.Cluster{}
	for _, h := range mr.hosts {
		metrics := mr.metrics[h.Name]
		logHostAndMetrics(mr.logger, h, metrics)
		host := domainHostFromModelHostAndMetrics(h, metrics)
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster
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
		Str("host", h.Name).
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
