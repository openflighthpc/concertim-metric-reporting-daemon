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

func dsm_for(hostname string) domain.DSM {
	return domain.DSM{
		GridName:    "unspecified",
		ClusterName: "unspecified",
		HostName:    hostname,
	}
}

func Test_NewRepoIsEmpty(t *testing.T) {
	repo := New(log.Logger)
	hosts := repo.GetAll()
	assert.Empty(t, hosts)
}

func Test_AddedHostCanBeRetrieved(t *testing.T) {
	tests := []struct {
		name  string
		hosts []domain.ReportedHost
	}{
		{
			name:  "Adding zero hosts",
			hosts: []domain.ReportedHost{},
		},
		{
			name: "Adding one host",
			hosts: []domain.ReportedHost{
				{
					Id:       "10",
					DSM:      dsm_for("comp10"),
					Reported: time.Now(),
					DMax:     10,
					Metrics:  []domain.ReportedMetric{},
				},
			},
		},
		{
			name: "Adding two hosts",
			hosts: []domain.ReportedHost{
				{
					Id:       "10",
					DSM:      dsm_for("comp10"),
					Reported: time.Now(),
					DMax:     10,
					Metrics:  []domain.ReportedMetric{},
				},
				{
					Id:       "20",
					DSM:      dsm_for("comp20"),
					Reported: time.Now(),
					DMax:     20,
					Metrics:  []domain.ReportedMetric{},
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
				h, ok := repo.GetHost(host.Id)
				assert.False(ok)
				assert.Equal(domain.ReportedHost{}, h)
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
		hosts   map[domain.HostId]domain.ReportedHost
		metrics map[domain.HostId][]domain.ReportedMetric
	}{
		{
			name: "Adding multiple metrics for a single host",
			hosts: map[domain.HostId]domain.ReportedHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: []domain.ReportedMetric{}},
			},
			metrics: map[domain.HostId][]domain.ReportedMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
					{Name: "temp", Value: "100", Units: "C", Slope: "both", DMax: 60, Type: "int32"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[domain.HostId]domain.ReportedHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: []domain.ReportedMetric{}},
				"20": {Id: "20", Reported: time.Now(), DMax: 20, Metrics: []domain.ReportedMetric{}},
			},
			metrics: map[domain.HostId][]domain.ReportedMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
				"20": {
					{Name: "power", Value: "100", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
			},
		},
		{
			name: "Adding different metrics for different hosts",
			hosts: map[domain.HostId]domain.ReportedHost{
				"10": {Id: "10", Reported: time.Now(), DMax: 10, Metrics: []domain.ReportedMetric{}},
				"20": {Id: "20", Reported: time.Now(), DMax: 20, Metrics: []domain.ReportedMetric{}},
			},
			metrics: map[domain.HostId][]domain.ReportedMetric{
				"10": {
					{Name: "power", Value: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
				},
				"20": {
					{Name: "power", Value: "100", Units: "W", Slope: "both", DMax: 60, Type: "int64"},
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
				assert.Contains(tt.hosts, host.Id)
				assert.Equal(
					sortMetrics(tt.metrics[host.Id]),
					sortMetrics(host.Metrics),
				)
			}
		})
	}
}

func sortMetrics(metrics []domain.ReportedMetric) []domain.ReportedMetric {
	sort.SliceStable(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})
	return metrics
}

func Test_AddingMetricForUnknownHostIsAnError(t *testing.T) {
	// Setup
	assert := assert.New(t)
	repo := New(log.Logger)
	host := domain.ReportedHost{Id: "01", Reported: time.Now(), DMax: 10, Metrics: []domain.ReportedMetric{}}
	metric := domain.ReportedMetric{Name: "power", Value: "10", Units: "W", Slope: "both", DMax: 60, Type: "int64"}

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
	repo := New(log.Logger)
	host := domain.ReportedHost{
		Id:       "01",
		DSM:      dsm_for("comp01"),
		Reported: time.Now(),
		DMax:     10,
		Metrics:  []domain.ReportedMetric{},
	}
	err := repo.PutHost(host)
	assert.NoError(err)

	// Preconditions
	hosts := repo.GetAll()
	assert.Len(hosts, 1)

	// Actions
	updatedHost := host
	updatedHost.Metrics = make([]domain.ReportedMetric, len(host.Metrics))
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
