package memory

import (
	"sync"
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
	"github.com/rs/zerolog"
)

type Memory struct {
	hosts   []Host
	metrics map[string]map[string]Metric
	mux     sync.Mutex
	logger  zerolog.Logger
}

func (m *Memory) PutHost(host domain.Host) error {
	m.logger.Debug().Str("host", host.Name).Msg("Putting host")
	m.mux.Lock()
	defer m.mux.Unlock()
	m.hosts = append(m.hosts, dbHostFromDomain(host))
	return nil
}

func (m *Memory) PutMetric(host domain.Host, metric domain.Metric) error {
	m.logger.Debug().Str("host", host.Name).Str("metric", metric.Name).Msg("Putting metric")
	m.mux.Lock()
	defer m.mux.Unlock()
	metrics, ok := m.metrics[host.Name]
	if !ok {
		metrics = make(map[string]Metric, 0)
		m.metrics[host.Name] = metrics
	}
	metrics[metric.Name] = dbMetricFromDomain(metric)
	return nil
}

func (m *Memory) GetAll() domain.Cluster {
	m.logger.Debug().Msg("Getting all data")
	cluster := domain.Cluster{}
	for _, h := range m.hosts {
		host := domainHostFromDb(h)
		m.logger.Debug().Str("host", host.Name).Msg("Found host")
		host.Metrics = make([]domain.Metric, 0)
		if metrics, ok := m.metrics[host.Name]; ok {
			for _, metric := range metrics {
				m.logger.Debug().Str("metric", metric.Name).Msg("Found metric")
				host.Metrics = append(host.Metrics, domainMetricFromDb(metric))
			}
		}
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster
}

func New(logger zerolog.Logger) *Memory {
	m := &Memory{
		hosts:   []Host{},
		metrics: map[string]map[string]Metric{},
		mux:     sync.Mutex{},
		logger:  logger.With().Str("component", "storage").Logger(),
	}
	addFakeData(m)
	return m
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
