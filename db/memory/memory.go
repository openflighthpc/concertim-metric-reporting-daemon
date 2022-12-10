package memory

import (
	"strings"
	"sync"
	"time"

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

func (m *Memory) PutMetric(host domain.Host, metric domain.Metric) error {
	m.logger.Debug().Str("host", host.Name).Str("metric", metric.Name).Msg("Putting metric")
	m.mux.Lock()
	defer m.mux.Unlock()
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
	addFakeData(m)
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

func addFakeData(m *Memory) {
	comp001 := domain.Host{Name: "comp001", Reported: time.Now().Add(-2 * time.Hour), TMax: 60 * time.Second, DMax: 60 * time.Second}
	comp002 := domain.Host{Name: "comp002", Reported: time.Now().Add(-3 * time.Hour), TMax: 60 * time.Second, DMax: 60 * time.Second}
	err := m.PutHost(comp001)
	if err != nil {
		m.logger.Warn().Err(err)
	}
	err = m.PutHost(comp002)
	if err != nil {
		m.logger.Warn().Err(err)
	}
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "foo",
			Val:    "foobar",
			Units:  "foos",
			Slope:  domain.MetricSlopeZero,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeString,
		},
	)
	if err != nil {
		m.logger.Warn().Err(err)
	}
	// Duplicate foo metric
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "foo",
			Val:    "FOOBAR",
			Units:  "FOOS",
			Slope:  domain.MetricSlopeZero,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeString,
		},
	)
	if err != nil {
		m.logger.Warn().Err(err)
	}
	err = m.PutMetric(comp001,
		domain.Metric{
			Name:   "bar",
			Val:    "12",
			Units:  "bars",
			Slope:  domain.MetricSlopeBoth,
			Tn:     0,
			TMax:   60 * time.Second,
			DMax:   60 * time.Second,
			Source: "MRAPI",
			Type:   domain.MetricTypeInt32,
		},
	)
	if err != nil {
		m.logger.Warn().Err(err)
	}
}
