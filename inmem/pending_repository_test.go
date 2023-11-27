package inmem

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

func dsm_for(hostname string) domain.DSM {
	return domain.DSM{
		GridName:    "unspecified",
		ClusterName: "unspecified",
		HostName:    hostname,
	}
}

func Test_NewRepoIsEmpty(t *testing.T) {
	repo := NewPendingRepository(log.Logger)
	hosts := repo.GetAll()
	assert.Empty(t, hosts)
}

func Test_AddedHostCanBeRetrieved(t *testing.T) {
	tests := []struct {
		name  string
		hosts []domain.PendingHost
	}{
		{
			name:  "Adding zero hosts",
			hosts: []domain.PendingHost{},
		},
		{
			name: "Adding one host",
			hosts: []domain.PendingHost{
				{
					Id:       "10",
					DSM:      dsm_for("comp10"),
					Reported: time.Now(),
					DMax:     10,
					Metrics:  map[domain.MetricName]domain.PendingMetric{},
				},
			},
		},
		{
			name: "Adding two hosts",
			hosts: []domain.PendingHost{
				{
					Id:       "10",
					DSM:      dsm_for("comp10"),
					Reported: time.Now(),
					DMax:     10,
					Metrics:  map[domain.MetricName]domain.PendingMetric{},
				},
				{
					Id:       "20",
					DSM:      dsm_for("comp20"),
					Reported: time.Now(),
					DMax:     20,
					Metrics:  map[domain.MetricName]domain.PendingMetric{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			repo := NewPendingRepository(log.Logger)

			// Preconditions
			hosts := repo.GetAll()
			assert.Empty(hosts)
			for _, host := range tt.hosts {
				h, ok := repo.GetHost(host.Id)
				assert.False(ok)
				assert.Equal(domain.PendingHost{}, h)
			}

			// Actions
			for _, host := range tt.hosts {
				err := repo.PutHost(host)
				assert.NoError(err)
			}

			// Assertions
			hosts = repo.GetAll()
			assert.Len(hosts, len(tt.hosts))
			for _, host := range hosts {
				assert.Contains(tt.hosts, host)
			}
			for _, host := range tt.hosts {
				h, ok := repo.GetHost(host.Id)
				assert.True(ok)
				assert.Equal(host, h)
			}
		})
	}
}

func Test_AddedMetricsCanBeRetrieved(t *testing.T) {
	tests := []struct {
		name    string
		hosts   map[domain.HostId]domain.PendingHost
		metrics map[domain.HostId][]domain.PendingMetric
	}{
		{
			name: "Adding multiple metrics for a single host",
			hosts: map[domain.HostId]domain.PendingHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: map[domain.MetricName]domain.PendingMetric{}},
			},
			metrics: map[domain.HostId][]domain.PendingMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", TTL: 60, Type: "int64"},
					{Name: "temp", Value: "100", Units: "C", Slope: "both", TTL: 60, Type: "int32"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[domain.HostId]domain.PendingHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: map[domain.MetricName]domain.PendingMetric{}},
				"20": {Id: "20", Reported: time.Now(), DMax: 20, Metrics: map[domain.MetricName]domain.PendingMetric{}},
			},
			metrics: map[domain.HostId][]domain.PendingMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", TTL: 60, Type: "int64"},
				},
				"20": {
					{Name: "power", Value: "100", Units: "W", Slope: "both", TTL: 60, Type: "int64"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			repo := NewPendingRepository(log.Logger)

			// Preconditions
			hosts := repo.GetAll()
			assert.Empty(hosts)

			// Actions
			for _, host := range tt.hosts {
				err := repo.PutHost(host)
				assert.NoError(err)
			}
			for hostName, metrics := range tt.metrics {
				host := tt.hosts[hostName]
				for _, metric := range metrics {
					err := repo.PutMetric(host, metric)
					assert.NoError(err)
				}
			}

			// Assertions
			hosts = repo.GetAll()
			assert.Len(hosts, len(tt.hosts))
			for _, host := range hosts {
				assert.Contains(tt.hosts, host.Id)
				// Convert to JSON to test without relying on key ordering.
				jsonMap1, _ := json.Marshal(tt.metrics[host.Id])
				jsonMap2, _ := json.Marshal(tt.metrics[host.Id])
				assert.Equal(jsonMap1, jsonMap2)
			}
		})
	}
}

func Test_AddingMetricForUnknownHostIsAnError(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := NewPendingRepository(log.Logger)
	host := domain.PendingHost{Id: "01", Reported: time.Now(), DMax: 10, Metrics: map[domain.MetricName]domain.PendingMetric{}}
	metric := domain.PendingMetric{Name: "power", Value: "10", Units: "W", Slope: "both", TTL: 60, Type: "int64"}

	// Preconditions
	hosts := repo.GetAll()
	assert.Empty(hosts)

	// Actions
	// Deliberately don't add host to the repo.
	err := repo.PutMetric(host, metric)

	// Assertions
	assert.Error(err)
	assert.ErrorIs(err, domain.ErrUnknownHost)
}

func Test_AddingHostUpdatesIfAlreadyThere(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := NewPendingRepository(log.Logger)
	host := domain.PendingHost{
		Id:       "01",
		DSM:      dsm_for("comp01"),
		Reported: time.Now(),
		DMax:     10,
		Metrics:  map[domain.MetricName]domain.PendingMetric{},
	}
	err := repo.PutHost(host)
	assert.NoError(err)

	// Preconditions
	hosts := repo.GetAll()
	assert.Len(hosts, 1)

	// Actions
	updatedHost := host
	updatedHost.Metrics = make(map[domain.MetricName]domain.PendingMetric)
	for k, v := range host.Metrics {
		updatedHost.Metrics[k] = v
	}
	updatedHost.DMax = host.DMax + 10
	err = repo.PutHost(updatedHost)

	// Assertions
	assert.NoError(err)
	hosts = repo.GetAll()
	assert.Len(hosts, 1)
	assert.Equal(updatedHost, hosts[0])
	assert.Equal(host.DMax+10, hosts[0].DMax)
}

func Test_UpdateLastProcessed(t *testing.T) {
	tests := []struct {
		name    string
		hosts   map[domain.HostId]domain.PendingHost
		metrics map[domain.HostId][]domain.PendingMetric
	}{
		{
			name: "UpdateLastProcessed when LastProcessed is nil",
			hosts: map[domain.HostId]domain.PendingHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: map[domain.MetricName]domain.PendingMetric{}},
				"20": {Id: "20", Reported: time.Now(), DMax: 20, Metrics: map[domain.MetricName]domain.PendingMetric{}},
			},
			metrics: map[domain.HostId][]domain.PendingMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", TTL: 60, Type: "int64"},
					{Name: "temp", Value: "100", Units: "C", Slope: "both", TTL: 60, Type: "int32"},
				},
				"20": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", TTL: 60, Type: "int64"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			// Setup
			repo := NewPendingRepository(log.Logger)
			for _, host := range tt.hosts {
				err := repo.PutHost(host)
				assert.NoError(err)
			}
			for hostName, metrics := range tt.metrics {
				host := tt.hosts[hostName]
				for _, metric := range metrics {
					err := repo.PutMetric(host, metric)
					assert.NoError(err)
				}
			}
			hostId := domain.HostId("10")
			metricName := domain.MetricName("power")

			// Preconditions
			host, ok := repo.GetHost(hostId)
			assert.True(ok, "expected host to be found")
			metric, ok := host.Metrics[metricName]
			assert.True(ok, "expected metric to be found")
			assert.Nil(metric.LastProcessed, "expected LastProcessed to be nil")

			// Actions
			now := time.Now()
			err := repo.UpdateLastProcessed(hostId, metricName, now)
			assert.NoError(err)

			// Assertions
			host, ok = repo.GetHost(hostId)
			assert.True(ok, "expected host to be found")
			metric, ok = host.Metrics[metricName]
			assert.True(ok, "expected metric to be found")
			assert.NotNil(metric.LastProcessed, "expected LastProcessed to have been updated")
			assert.Equal(now, *metric.LastProcessed, "expected LastProcessed to have been updated")

			// Check no other metrics have been updated.
			hosts := repo.GetAll()
			for _, host := range hosts {
				for _, metric := range host.Metrics {
					if host.Id == hostId && metric.Name == string(metricName) {
						continue
					}
					assert.Nil(metric.LastProcessed, "expected other metrics not to have been updated")
				}
			}
		})
	}
}
