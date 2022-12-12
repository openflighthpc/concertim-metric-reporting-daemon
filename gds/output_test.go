package gds

import (
	"fmt"
	"io"
	"os"
	"strconv"
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

type fakeClock struct{}

func (fakeClock) Now() time.Time {
	i, err := strconv.ParseInt("1670856697", 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Unix(i, 0)
}

func Test_GeneratedXMLIsValid(t *testing.T) {
	tests := []struct {
		name    string
		cluster domain.Cluster
		golden  string
	}{
		{
			name:    "generates correct XML for empty cluster",
			cluster: domain.Cluster{Hosts: []domain.Host{}},
			golden:  "empty_cluster",
		},
		{
			name:    "generates correct XML for cluster with hosts without metrics",
			cluster: clusterWithoutMetrics(),
			golden:  "cluster_without_metrics",
		},
		{
			name:    "generates correct XML for cluster with hosts with metrics",
			cluster: clusterWithMetrics(),
			golden:  "cluster_with_metrics",
		},
	}
	outputGenerator := newOutputGenerator(fakeClock{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			// Action
			output, err := outputGenerator.generate(tt.cluster)

			// Assertions
			assert.NoError(err)
			assert.Equal(goldenValue(t, tt.golden), string(output))
		})
	}
}

func clusterWithoutMetrics() domain.Cluster {
	return domain.Cluster{Hosts: []domain.Host{
		{
			Name:     "comp10",
			Reported: fakeClock{}.Now(),
			TMax:     10 * time.Second,
			DMax:     10 * time.Second,
			Metrics:  []domain.Metric{},
		},
		{
			Name:     "comp20",
			Reported: fakeClock{}.Now(),
			TMax:     20 * time.Second,
			DMax:     20 * time.Second,
			Metrics:  []domain.Metric{},
		},
	}}
}

func clusterWithMetrics() domain.Cluster {
	cluster := clusterWithoutMetrics()
	var hosts []domain.Host
	for i, host := range cluster.Hosts {
		host.Metrics = append(host.Metrics, buildMetrics(i+1)...)
		hosts = append(hosts, host)
	}
	cluster.Hosts = hosts
	return cluster
}

func buildMetrics(i int) []domain.Metric {
	powerMetric := domain.Metric{
		Name:   "power",
		Val:    fmt.Sprintf("%d", i*10),
		Units:  "W",
		Slope:  "both",
		Tn:     0,
		TMax:   60 * time.Second,
		DMax:   60 * time.Second,
		Source: "mrapi",
		Type:   domain.MetricTypeDouble,
	}
	tempMetric := domain.Metric{
		Name:   "temp",
		Val:    fmt.Sprintf("%d", i*20),
		Units:  "C",
		Slope:  "both",
		Tn:     0,
		TMax:   120 * time.Second,
		DMax:   120 * time.Second,
		Source: "mrapi",
		Type:   domain.MetricTypeFloat,
	}
	return []domain.Metric{powerMetric, tempMetric}
}

func goldenValue(t *testing.T, goldenFile string) string {
	t.Helper()
	goldenPath := "testdata/" + goldenFile + ".golden"

	f, err := os.Open(goldenPath)
	defer func() { _ = f.Close() }()
	if err != nil {
		t.Fatalf("Error opening file %s: %s", goldenPath, err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Error opening file %s: %s", goldenPath, err)
	}
	return string(content)
}
