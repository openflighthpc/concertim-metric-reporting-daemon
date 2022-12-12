package memory

import (
	"os"
	"sort"
	"testing"
	"time"

	domain "github.com/alces-flight/concertim-mrapi/domain"
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
	cluster := repo.GetAll()
	assert.Empty(t, cluster.Hosts)
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
					Name:     "comp10",
					Reported: time.Now(),
					TMax:     10,
					DMax:     10,
					Metrics:  []domain.Metric{},
				},
			},
		},
		{
			name: "Adding two hosts",
			hosts: []domain.Host{
				{
					Name:     "comp10",
					Reported: time.Now(),
					TMax:     10,
					DMax:     10,
					Metrics:  []domain.Metric{},
				},
				{
					Name:     "comp20",
					Reported: time.Now(),
					TMax:     20,
					DMax:     20,
					Metrics:  []domain.Metric{},
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
			cluster := repo.GetAll()
			assert.Empty(cluster.Hosts)

			// Actions
			for _, host := range tt.hosts {
				err := repo.PutHost(host)
				assert.NoError(err)
			}

			// Assertions
			cluster = repo.GetAll()
			assert.Len(cluster.Hosts, len(tt.hosts))
			for _, host := range cluster.Hosts {
				assert.Contains(tt.hosts, host)
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
				"comp10": {Name: "comp10", Reported: time.Now(), TMax: 10, DMax: 10, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": []domain.Metric{
					domain.Metric{Name: "power", Val: "10", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"},
					domain.Metric{Name: "temp", Val: "100", Units: "C", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int32"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[string]domain.Host{
				"comp10": {Name: "comp10", Reported: time.Now(), TMax: 10, DMax: 10, Metrics: []domain.Metric{}},
				"comp20": {Name: "comp20", Reported: time.Now(), TMax: 20, DMax: 20, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": []domain.Metric{
					domain.Metric{Name: "power", Val: "10", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"},
				},
				"comp20": []domain.Metric{
					domain.Metric{Name: "power", Val: "100", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[string]domain.Host{
				"comp10": {Name: "comp10", Reported: time.Now(), TMax: 10, DMax: 10, Metrics: []domain.Metric{}},
				"comp20": {Name: "comp20", Reported: time.Now(), TMax: 20, DMax: 20, Metrics: []domain.Metric{}},
			},
			metrics: map[string][]domain.Metric{
				"comp10": []domain.Metric{
					domain.Metric{Name: "power", Val: "10", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"},
				},
				"comp20": []domain.Metric{
					domain.Metric{Name: "power", Val: "100", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"},
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
			cluster := repo.GetAll()
			assert.Empty(cluster.Hosts)

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
			cluster = repo.GetAll()
			assert.Len(cluster.Hosts, len(tt.hosts))
			for _, host := range cluster.Hosts {
				assert.Contains(tt.hosts, host.Name)
				assert.Equal(
					sortMetrics(tt.metrics[host.Name]),
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
	host := domain.Host{Name: "comp01", Reported: time.Now(), TMax: 10, DMax: 10, Metrics: []domain.Metric{}}
	metric := domain.Metric{Name: "power", Val: "10", Units: "W", Slope: "both", Tn: 0, TMax: 60, DMax: 60, Source: "mrapi", Type: "int64"}

	// Preconditions
	cluster := repo.GetAll()
	assert.Empty(cluster.Hosts)

	// Actions
	// Deliberately don't add host to the repo.
	err := repo.PutMetric(host, metric)

	// Assertions
	assert.Error(err)
	assert.ErrorAs(err, &domain.UnknownHost{})
}

func Test_AddingHostUpdatesIfAlreadyThere(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := New(log.Logger)
	host := domain.Host{Name: "comp01", Reported: time.Now(), TMax: 10, DMax: 10, Metrics: []domain.Metric{}}
	err := repo.PutHost(host)
	assert.NoError(err)

	// Preconditions
	cluster := repo.GetAll()
	assert.Len(cluster.Hosts, 1)

	// Actions
	updatedHost := host
	updatedHost.Metrics = make([]domain.Metric, len(host.Metrics))
	copy(updatedHost.Metrics, host.Metrics)
	updatedHost.TMax = host.TMax + 10
	err = repo.PutHost(updatedHost)

	// Assertions
	assert.NoError(err)
	cluster = repo.GetAll()
	assert.Len(cluster.Hosts, 1)
	assert.Equal(updatedHost, cluster.Hosts[0])
	assert.Equal(host.TMax+10, cluster.Hosts[0].TMax)
}
