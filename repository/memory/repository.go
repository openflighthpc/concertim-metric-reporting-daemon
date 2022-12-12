package memory

import (
	"strings"
	"sync"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/rs/zerolog"
)

// MemoryRepo is an in-memory repository.  All data it stores will be lost
// when the process exits.
type MemoryRepo struct {
	hosts   map[string]HostModel
	metrics map[string]map[string]MetricModel
	mux     sync.Mutex
	logger  zerolog.Logger
}

func (mr *MemoryRepo) PutHost(host domain.Host) error {
	mr.logger.Debug().Str("host", host.Name).Msg("Putting host")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.hosts[host.Name] = conv.ModelFromDomainHost(host)
	return nil
}

func (mr *MemoryRepo) isHostStored(host domain.Host) bool {
	_, ok := mr.hosts[host.Name]
	return ok
}

func (mr *MemoryRepo) PutMetric(host domain.Host, metric domain.Metric) error {
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

func (mr *MemoryRepo) GetAll() domain.Cluster {
	mr.logger.Debug().Msg("Getting all data")
	cluster := domain.Cluster{}
	for _, h := range mr.hosts {
		metrics := mr.metrics[h.Name]
		logHostAndMetrics(mr.logger, h, metrics)
		host := DomainHostFromModelHostAndMetrics(h, metrics)
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster
}

func New(logger zerolog.Logger) *MemoryRepo {
	mr := &MemoryRepo{
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
