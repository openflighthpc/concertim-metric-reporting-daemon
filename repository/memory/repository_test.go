package memory

import (
	"os"
	"sort"
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

func Test_NewRepoIsEmpty(t *testing.T) {
	repo := New(log.Logger)
	hosts := repo.GetAll()
	assert.Empty(t, hosts)
}

func Test_AddedHostCanBeRetrieved(t *testing.T) {
	tests := []struct {
		name  string
		hosts []domain.Host
	}{
		{
			name:  "Adding zero hosts",
			hosts: []domain.Host{},
		},
		{
			name: "Adding one host",
			hosts: []domain.Host{
				{
					DeviceName: "comp10",
					Reported:   time.Now(),
					DMax:       10,
					Metrics:    []domain.Metric{},
				},
			},
		},
		{
			name: "Adding two hosts",
			hosts: []domain.Host{
				{
					DeviceName: "comp10",
					Reported:   time.Now(),
					DMax:       10,
					Metrics:    []domain.Metric{},
				},
				{
					DeviceName: "comp20",
					Reported:   time.Now(),
					DMax:       20,
					Metrics:    []domain.Metric{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			repo := New(log.Logger)

			// Preconditions
			hosts := repo.GetAll()
			assert.Empty(hosts)
			for _, host := range tt.hosts {
				h, ok := repo.GetHost(host.DeviceName)
				assert.False(ok)
				assert.Equal(domain.Host{}, h)
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
				h, ok := repo.GetHost(host.DeviceName)
				assert.True(ok)
				assert.Equal(host, h)
			}
		})
	}
}

func Test_AddedMetricsCanBeRetrieved(t *testing.T) {
	tests := []struct {
		name    string
		hosts   map[string]domain.Host
		metrics map[string][]domain.Metric
	}{
		{
			name: "Adding multiple metrics for a single host",
			hosts: map[string]domain.Host{
				"comp10": {DeviceName: "comp10", Reported: time.Now(), DMax: 10, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": {
					{Name: "power", Val: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
					{Name: "temp", Val: "100", Units: "C", Slope: "both", DMax: 60, Type: "int32"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[string]domain.Host{
				"comp10": {DeviceName: "comp10", Reported: time.Now(), DMax: 10, Metrics: []domain.Metric{}},
				"comp20": {DeviceName: "comp20", Reported: time.Now(), DMax: 20, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": {
					{Name: "power", Val: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
				"comp20": {
					{Name: "power", Val: "100", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[string]domain.Host{
				"comp10": {DeviceName: "comp10", Reported: time.Now(), DMax: 10, Metrics: []domain.Metric{}},
				"comp20": {DeviceName: "comp20", Reported: time.Now(), DMax: 20, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": {
					{Name: "power", Val: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
				"comp20": {
					{Name: "power", Val: "100", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			repo := New(log.Logger)

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
				assert.Contains(tt.hosts, host.DeviceName)
				assert.Equal(
					sortMetrics(tt.metrics[host.DeviceName]),
					sortMetrics(host.Metrics),
				)
			}
		})
	}
}

func sortMetrics(metrics []domain.Metric) []domain.Metric {
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics
}

func Test_AddingMetricForUnknownHostIsAnError(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := New(log.Logger)
	host := domain.Host{DeviceName: "comp01", Reported: time.Now(), DMax: 10, Metrics: []domain.Metric{}}
	metric := domain.Metric{Name: "power", Val: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"}

	// Preconditions
	hosts := repo.GetAll()
	assert.Empty(hosts)

	// Actions
	// Deliberately don't add host to the repo.
	err := repo.PutMetric(host, metric)

	// Assertions
	assert.Error(err)
	assert.ErrorIs(err, domain.UnknownHost)
}

func Test_AddingHostUpdatesIfAlreadyThere(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := New(log.Logger)
	host := domain.Host{DeviceName: "comp01", Reported: time.Now(), DMax: 10, Metrics: []domain.Metric{}}
	err := repo.PutHost(host)
	assert.NoError(err)

	// Preconditions
	hosts := repo.GetAll()
	assert.Len(hosts, 1)

	// Actions
	updatedHost := host
	updatedHost.Metrics = make([]domain.Metric, len(host.Metrics))
	copy(updatedHost.Metrics, host.Metrics)
	updatedHost.DMax = host.DMax + 10
	err = repo.PutHost(updatedHost)

	// Assertions
	assert.NoError(err)
	hosts = repo.GetAll()
	assert.Len(hosts, 1)
	assert.Equal(updatedHost, hosts[0])
	assert.Equal(host.DMax+10, hosts[0].DMax)
}
