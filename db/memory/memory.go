package memory

import (
	"strings"
	"sync"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/rs/zerolog"
)

type Memory struct {
	hosts   []HostModel
	metrics map[string]map[string]MetricModel
	mux     sync.Mutex
	logger  zerolog.Logger
}

func (m *Memory) PutHost(host domain.Host) error {
	m.logger.Debug().Str("host", host.Name).Msg("Putting host")
	m.mux.Lock()
	defer m.mux.Unlock()
	m.hosts = append(m.hosts, conv.ModelFromDomainHost(host))
	return nil
}

func (m *Memory) isHostStored(host domain.Host) bool {
	for _, dbHost := range m.hosts {
		if dbHost.Name == host.Name {
			return true
		}
	}
	return false
}

func (m *Memory) PutMetric(host domain.Host, metric domain.Metric) error {
	m.logger.Debug().Str("host", host.Name).Str("metric", metric.Name).Msg("Putting metric")
	m.mux.Lock()
	defer m.mux.Unlock()
	if !m.isHostStored(host) {
		return domain.UnknownHost{Host: host}
	}
	metrics, ok := m.metrics[host.Name]
	if !ok {
		metrics = make(map[string]MetricModel, 0)
		m.metrics[host.Name] = metrics
	}
	metrics[metric.Name] = conv.ModelFromDomainMetric(metric)
	return nil
}

func (m *Memory) GetAll() domain.Cluster {
	m.logger.Debug().Msg("Getting all data")
	cluster := domain.Cluster{}
	for _, h := range m.hosts {
		metrics := m.metrics[h.Name]
		logHostAndMetrics(m.logger, h, metrics)
		host := DomainHostFromModelHostAndMetrics(h, metrics)
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster
}

func New(logger zerolog.Logger) *Memory {
	m := &Memory{
		hosts:   []HostModel{},
		metrics: map[string]map[string]MetricModel{},
		mux:     sync.Mutex{},
		logger:  logger.With().Str("component", "storage").Logger(),
	}
	return m
}

func logHostAndMetrics(log zerolog.Logger, h HostModel, metrics map[string]MetricModel) {
	log.Debug().
		Str("host", h.Name).
		Int("metric.count", len(metrics)).
		Func(func(e *zerolog.Event) {
			metricNames := []string{}
			for _, m := range metrics {
				metricNames = append(metricNames, m.Name)
			}
			e.Str("metric.names", strings.Join(metricNames, ","))
		}).
		Msg("Found host")
}
