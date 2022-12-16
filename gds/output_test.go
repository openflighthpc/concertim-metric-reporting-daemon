package gds

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html/charset"
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
func (fakeClock) Since(t time.Time) time.Duration {
	return 32 * time.Second
}

func Test_GeneratedXMLIsCorrect(t *testing.T) {
	tests := []struct {
		name   string
		hosts  []domain.Host
		golden string
	}{
		{
			name:   "generates correct XML for empty cluster",
			hosts:  []domain.Host{},
			golden: "empty_cluster",
		},
		{
			name:   "generates correct XML for cluster with hosts without metrics",
			hosts:  clusterWithoutMetrics(),
			golden: "cluster_without_metrics",
		},
		{
			name:   "generates correct XML for cluster with hosts with metrics",
			hosts:  clusterWithMetrics(),
			golden: "cluster_with_metrics",
		},
		{
			name:   "escapes XML correctly",
			hosts:  clusterWithXML(),
			golden: "escaped_xml",
		},
	}
	config := config.GDS{
		ClusterName:  "unspecified",
		MetricSource: "ct-metric-reporting-daemon",
	}
	outputGenerator, err := newOutputGenerator(fakeClock{}, config)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action
			output, err := outputGenerator.generate(tt.hosts)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, goldenValue(t, tt.golden), string(output))
			assert.NoError(t, parseXML(output))
		})
	}
}

func parseXML(theXML []byte) error {
	reader := bytes.NewReader(theXML)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(new(interface{}))
	return err
}

func clusterWithoutMetrics() []domain.Host {
	return []domain.Host{
		{
			DeviceName: "comp10",
			DSMName:    "comp10.cluster.local",
			Reported:   fakeClock{}.Now(),
			DMax:       10 * time.Second,
			Metrics:    []domain.Metric{},
		},
		{
			DeviceName: "comp20",
			DSMName:    "comp20.cluster.local",
			Reported:   fakeClock{}.Now(),
			DMax:       20 * time.Second,
			Metrics:    []domain.Metric{},
		},
	}
}

func clusterWithMetrics() []domain.Host {
	origHosts := clusterWithoutMetrics()
	newHosts := make([]domain.Host, 0, len(origHosts))
	log.Printf("origHosts: %#v", origHosts)
	for i, host := range origHosts {
		host.Metrics = append(host.Metrics, buildMetrics(i+1)...)
		newHosts = append(newHosts, host)
	}
	return newHosts
}

func buildMetrics(i int) []domain.Metric {
	powerMetric := domain.Metric{
		Name:     "power",
		Val:      fmt.Sprintf("%d", i*10),
		Units:    "W",
		Slope:    "both",
		DMax:     60 * time.Second,
		Reported: fakeClock{}.Now(),
		Type:     domain.MetricTypeDouble,
	}
	tempMetric := domain.Metric{
		Name:     "temp",
		Val:      fmt.Sprintf("%d", i*20),
		Units:    "C",
		Slope:    "both",
		DMax:     120 * time.Second,
		Reported: fakeClock{}.Now(),
		Type:     domain.MetricTypeFloat,
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

func clusterWithXML() []domain.Host {
	return []domain.Host{
		{
			DeviceName: "\"</HOST>",
			DSMName:    "\"</HOST>.cluster.local",
			Reported:   fakeClock{}.Now(),
			DMax:       10 * time.Second,
			Metrics: []domain.Metric{{
				Name:  "\"</NAME>",
				Val:   "\"</VAL>",
				Units: "\"</UNITS>",
			}},
		},
	}
}
