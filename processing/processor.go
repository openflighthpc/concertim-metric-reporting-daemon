// Package processing collects stuff about processing.
package processing

import (
	"reflect"
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

// Processor is a struct for processing the metrics retrieved from Ganglia.
//
// It produces three different views of the metrics:
//
//  1. A list of unique metrics.
//  2. For each unique metric, a list of which devices are currently reporting
//     that metric.
//  3. For each host a map from metric name to that metric.
//
// These views are currently, recorded in memcache by Recorder.
type Processor struct {
	logger zerolog.Logger
}

// NewProcessor returns a new *Processor.
func NewProcessor(logger zerolog.Logger) *Processor {
	return &Processor{
		logger: logger.With().Str("component", "processor").Logger(),
	}
}

// Process processes the provided ganglia metrics and returns a *Result struct
// containing the results of the processing run.
//
// Any grids with a name other than "unspecified" or clusters with a name other
// than the GDS is configured with are ignored. This allows (with suitable
// configuration) Ganglia's gmond to run and collect metrics for localhost
// without storing them in memcache.
func (p *Processor) Process(hosts []*domain.ProcessedHost) (*Result, error) {
	result := NewResult()
	p.logger.Debug().Int("count", len(hosts)).Msg("processing hosts")
	for _, host := range hosts {
		p.logger.Debug().
			Str("host", host.DSM.HostName).
			Int("count", len(host.Metrics)).
			Msg("processing metrics")
		for _, metric := range host.Metrics {
			p.logger.Debug().
				Str("host", host.DSM.HostName).
				Str("metric", metric.Name).
				Msg("processing metric")
			result.AddMetric(host, &metric)
		}
		result.AddHost(host)
	}
	logProcessResults(p.logger, result)
	return result, nil
}

func logProcessResults(logger zerolog.Logger, result *Result) {
	logger.Info().
		Dict("processed", zerolog.Dict().
			Int("hosts", len(result.Hosts)).
			Int("metrics", result.numMetrics).
			Int("unique metrics", len(result.UniqueMetrics))).
		Dict("ignored", zerolog.Dict().
			Int("hosts", result.numIgnoredHosts)).
		Msg("completed")
}

// Result contains the result of the processing run for a single set of
// metrics.
//
// See the documentation of Processor for details of the views it provides.
type Result struct {
	// HostsByMetric is a map from a metric's name to a list of hosts that
	// currently have a fresh value for that metric.
	HostsByMetric map[domain.MetricName][]*domain.ProcessedHost

	// UniqueMetrics is a set of unique metrics by name.
	UniqueMetrics map[domain.MetricName]*domain.UniqueMetric

	// Hosts is a slice of Host containing their processed metrics.
	Hosts []*domain.ProcessedHost

	// numMetrics keeps track of the total metrics processed for logging.  Other
	// logged information is derived from the other attributes.
	numMetrics      int
	numIgnoredHosts int
}

var _ ResultRepo = (*Result)(nil)

// NewResult returns a new empty *Result.
func NewResult() *Result {
	return &Result{
		HostsByMetric: map[domain.MetricName][]*domain.ProcessedHost{},
		UniqueMetrics: map[domain.MetricName]*domain.UniqueMetric{},
	}
}

func (r *Result) AddHost(host *domain.ProcessedHost) {
	r.Hosts = append(r.Hosts, host)
	if len(host.Metrics) > 0 {
		now := time.Now()
		host.Mtime = &now
	}
}

// AddMetric adds the given Metric for the given MemcacheKey.  MemcacheKey
// should be the memcache key for the host that reported the metric.
func (r *Result) AddMetric(host *domain.ProcessedHost, metric *domain.ProcessedMetric) {
	metricName := domain.MetricName(metric.Name)
	host.Metrics[metricName] = *metric
	hosts, ok := r.HostsByMetric[metricName]
	if !ok {
		hosts = make([]*domain.ProcessedHost, 0)
		r.HostsByMetric[metricName] = hosts
	}
	r.HostsByMetric[metricName] = append(hosts, host)
	um, found := r.UniqueMetrics[metricName]
	if !found {
		um = uniqueMetricFromMetric(*metric)
		r.UniqueMetrics[metricName] = um
	}
	adjustMinMax(um, *metric)
	r.numMetrics++
}

func uniqueMetricFromMetric(src domain.ProcessedMetric) *domain.UniqueMetric {
	var dst domain.UniqueMetric
	dst.Datatype = src.Datatype
	dst.Name = src.Name
	dst.Nature = src.Nature
	dst.Units = src.Units
	adjustMinMax(&dst, src)
	return &dst
}

func adjustMinMax(unique *domain.UniqueMetric, metric domain.ProcessedMetric) {
	// XXX Add some logging of what's going on.  Especially for the error cases.
	switch metric.Datatype {
	case "int8", "int16", "int32":
		i, err := strconv.Atoi(metric.Value)
		if err != nil {
			return
		}
		if unique.Min == nil {
			unique.Min = i
		} else {
			minVal := reflect.ValueOf(unique.Min)
			if minVal.CanInt() {
				if int64(i) < minVal.Int() {
					unique.Min = i
				}
			}
		}
		if unique.Max == nil {
			unique.Max = i
		} else {
			maxVal := reflect.ValueOf(unique.Max)
			if maxVal.CanInt() {
				if int64(i) > maxVal.Int() {
					unique.Max = i
				}
			}
		}
	}
}
