// Package processing collects stuff about processing.
package processing

import (
	"reflect"
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/retrieval"
	"github.com/pkg/errors"
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
	clusterName string
	dsmRepo     domain.DataSourceMapRepository
	gridName    string
	logger      zerolog.Logger
}

// NewProcessor returns a new *Processor.
func NewProcessor(logger zerolog.Logger, dsmRepo domain.DataSourceMapRepository, gridName, clusterName string) *Processor {
	return &Processor{
		clusterName: clusterName,
		dsmRepo:     dsmRepo,
		gridName:    gridName,
		logger:      logger.With().Str("component", "processor").Logger(),
	}
}

// Process processes the provided ganglia metrics and returns a *Result struct
// containing the results of the processing run.
//
// Any grids with a name other than "unspecified" or clusters with a name other
// than the GDS is configured with are ignored. This allows (with suitable
// configuration) Ganglia's gmond to run and collect metrics for localhost
// without storing them in memcache.
func (p *Processor) Process(hosts []retrieval.Host) (*Result, error) {
	now := time.Now().Unix()
	err := p.dsmRepo.Update()
	if err != nil {
		p.logger.Warn().Err(err).Msg("using stale DSM data")
	}
	result := NewResult()
	p.logger.Debug().Int("count", len(hosts)).Msg("processing hosts")
	for _, gHost := range hosts {
		dsm := domain.DSM{
			GridName:    p.gridName,
			ClusterName: p.clusterName,
			HostName:    gHost.Name,
		}
		hostId, ok := p.dsmRepo.GetHostId(dsm)
		if !ok {
			p.logger.Debug().
				Str("host", gHost.Name).
				Stringer("dsm", dsm).
				Msg("ignoring")
			result.numIgnoredHosts++
			continue
		}
		p.logger.Debug().
			Str("host", gHost.Name).
			Int("count", len(gHost.Metrics)).
			Msg("processing metrics")
		ctHost := domain.ProcessedHost{
			Id:      hostId,
			DSM:     dsm,
			Metrics: make(map[domain.MetricName]domain.ProcessedMetric),
		}
		for _, gMetric := range gHost.Metrics {
			p.logger.Debug().
				Str("host", gHost.Name).
				Str("metric", gMetric.Name).
				Msg("processing metric")
			ctMetric, err := MetricFromGanglia(now, gMetric)
			if err != nil {
				p.logger.Warn().Err(err).Msg("failed")
				continue
			}
			// XXX Combine into single call. `result.AddMetric` or
			// perhaps `result.AddHost` and then
			// `result.AddMetric`.  I.e., use an interface not
			// direct access to maps.
			ctHost.Metrics[domain.MetricName(ctMetric.Name)] = ctMetric
			result.AddMetric(&ctHost, &ctMetric)
		}
		if len(ctHost.Metrics) > 0 {
			now := time.Now()
			ctHost.Mtime = &now
		}
		result.Hosts = append(result.Hosts, &ctHost)
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

// NewResult returns a new empty *Result.
func NewResult() *Result {
	return &Result{
		HostsByMetric: map[domain.MetricName][]*domain.ProcessedHost{},
		UniqueMetrics: map[domain.MetricName]*domain.UniqueMetric{},
	}
}

// AddMetric adds the given Metric for the given MemcacheKey.  MemcacheKey
// should be the memcache key for the host that reported the metric.
func (r *Result) AddMetric(host *domain.ProcessedHost, metric *domain.ProcessedMetric) {
	metricName := domain.MetricName(metric.Name)
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

// MetricFromGanglia parses the given Ganglia metric into a format suitable
// for processing.
//
// The formats are largely similar with the following changes:
//
// Slope is replaced with a simplified Nature.
// TN is replaced with a more easily consumed Timestamp.
// TMAX is replaced with a more easily consumed Stale.
func MetricFromGanglia(now int64, src retrieval.Metric) (domain.ProcessedMetric, error) {
	var dst domain.ProcessedMetric
	tn, err := strconv.Atoi(src.TN)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid TN")
	}
	tmax, err := strconv.Atoi(src.TMax)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid TMAX")
	}
	dmax, err := strconv.Atoi(src.DMax)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid DMAX")
	}

	var stale bool
	persistent := dmax == 0
	if persistent {
		stale = tn > tmax*2
	} else {
		stale = tn > dmax
	}

	nature := "volatile"
	if src.Type == "string" || src.Type == "timestamp" {
		nature = "string_and_time"
	} else if src.Slope == "zero" {
		nature = "constant"
	}

	dst.Name = src.Name
	dst.Datatype = src.Type
	dst.Units = src.Units
	dst.Source = src.Source
	dst.Value = src.Val
	dst.Nature = nature
	dst.Dmax = dmax
	dst.Timestamp = now - int64(tn)
	dst.Stale = stale

	return dst, nil
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
