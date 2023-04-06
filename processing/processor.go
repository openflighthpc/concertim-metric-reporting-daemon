package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/retrieval"
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
	dsmRepo *DSMRepo
	logger  zerolog.Logger
}

// NewProcessor returns a new *Processor.
func NewProcessor(logger zerolog.Logger, dsmRepo *DSMRepo) *Processor {
	return &Processor{
		dsmRepo: dsmRepo,
		logger:  logger.With().Str("component", "processor").Logger(),
	}
}

// Process processes the provided ganglia metrics and returns a *Result struct
// containing the results of the processing run.
//
// Any grids or clusters with a name other than "unspecified" are ignored.
// This allows (with suitable configuration) Ganglia's gmond to run and
// collect metrics for localhost without storing them in memcache.
func (p *Processor) Process(grids []retrieval.Grid) (*Result, error) {
	now := time.Now().Unix()
	err := p.dsmRepo.Update()
	if err != nil {
		p.logger.Warn().Err(err).Msg("using stale DSM data")
	}
	result := NewResult()
	p.logger.Debug().Int("count", len(grids)).Msg("processing grids")
	for _, gGrid := range grids {
		if gGrid.Name != "unspecified" {
			p.logger.Warn().Str("grid", gGrid.Name).Msg("ignoring")
			result.numIgnoredGrids++
			continue
		}
		p.logger.Debug().
			Str("grid", gGrid.Name).
			Int("count", len(gGrid.Clusters)).
			Msg("processing clusters")
		for _, gCluster := range gGrid.Clusters {
			if gCluster.Name != "unspecified" {
				p.logger.Warn().Str("cluster", gCluster.Name).Msg("ignoring")
				result.numIgnoredClusters++
				continue
			}
			p.logger.Debug().
				Str("grid", gGrid.Name).
				Str("cluster", gCluster.Name).
				Int("count", len(gCluster.Hosts)).
				Msg("processing hosts")
			for _, gHost := range gCluster.Hosts {
				dsm := DSM{
					GridName:    gGrid.Name,
					ClusterName: gCluster.Name,
					HostName:    gHost.Name,
				}
				memcacheKey, ok := p.dsmRepo.Get(dsm)
				if !ok {
					p.logger.Debug().
						Str("host", gHost.Name).
						Stringer("dsm", dsm).
						Msg("ignoring")
					result.numIgnoredHosts++
					continue
				}
				p.logger.Debug().
					Str("grid", gGrid.Name).
					Str("cluster", gCluster.Name).
					Str("host", gHost.Name).
					Int("count", len(gHost.Metrics)).
					Msg("processing metrics")
				ctHost := Host{
					Name:        gHost.Name,
					MemcacheKey: memcacheKey,
					Metrics:     make(map[MetricName]Metric),
				}
				for _, gMetric := range gHost.Metrics {
					p.logger.Debug().
						Str("grid", gGrid.Name).
						Str("cluster", gCluster.Name).
						Str("host", gHost.Name).
						Str("metric", gMetric.Name).
						Msg("processing metric")

					ctMetric, err := MetricFromGanglia(now, gMetric)
					if err != nil {
						p.logger.Err(err).Msg("failed")
						continue
					}
					ctHost.Metrics[MetricName(ctMetric.Name)] = ctMetric
					result.AddMetric(MemcacheKey(memcacheKey), ctMetric)
				}
				if len(ctHost.Metrics) > 0 {
					now := time.Now()
					ctHost.Mtime = &now
				}
				result.Hosts = append(result.Hosts, ctHost)
			}
		}
	}
	p.logger.Info().
		Dict("processed", zerolog.Dict().
			Int("hosts", len(result.Hosts)).
			Int("metrics", result.numMetrics).
			Int("unique metrics", len(result.UniqueMetrics))).
		Dict("ignored", zerolog.Dict().
			Int("grids", result.numIgnoredGrids).
			Int("clusters", result.numIgnoredClusters).
			Int("hosts", result.numIgnoredHosts)).
		Msg("completed")
	return result, nil
}

// MetricName exists to document some function signatures.
type MetricName string

// Result contains the result of the processing run for a single set of
// metrics.
//
// See the documentation of Processor for details of the views it provides.
type Result struct {
	// HostsByMetric is a map from a metric's name to a list of hosts that
	// currently have a fresh value for that metric.
	HostsByMetric map[MetricName][]MemcacheKey `json:"hosts_by_metric"`

	// UniqueMetrics is a set of unique metrics by name.
	UniqueMetrics map[MetricName]Metric

	// Hosts is a slice of Host containing their processed metrics.
	Hosts []Host `json:"hosts"`

	// numMetrics keeps track of the total metrics processed for logging.  Other
	// logged information is derived from the other attributes.
	numMetrics         int
	numIgnoredGrids    int
	numIgnoredClusters int
	numIgnoredHosts    int
}

// NewResult returns a new empty *Result.
func NewResult() *Result {
	return &Result{
		HostsByMetric: map[MetricName][]MemcacheKey{},
		UniqueMetrics: map[MetricName]Metric{},
	}
}

// AddMetric adds the given Metric for the given MemcacheKey.  MemcacheKey
// should be the memcache key for the host that reported the metric.
func (r *Result) AddMetric(mckey MemcacheKey, metric Metric) {
	metricName := MetricName(metric.Name)
	hosts, ok := r.HostsByMetric[metricName]
	if !ok {
		hosts = make([]MemcacheKey, 0)
		r.HostsByMetric[metricName] = hosts
	}
	r.HostsByMetric[metricName] = append(hosts, mckey)
	r.UniqueMetrics[metricName] = metric
	r.numMetrics++
}

// MarshalJSON implements the encoding/json.Marshaler interface.
//
// It is used here to provide a custom serialisation for UniqueMetrics.
func (r *Result) MarshalJSON() ([]byte, error) {
	uniqMetricNames := make([]Metric, 0, len(r.UniqueMetrics))
	for _, val := range r.UniqueMetrics {
		uniqMetricNames = append(uniqMetricNames, val)
	}
	return json.Marshal(&struct {
		HostsByMetric map[MetricName][]MemcacheKey `json:"hosts_by_metric"`
		Hosts         []Host                       `json:"hosts"`
		UniqueMetrics []Metric                     `json:"unique_metrics"`
	}{
		HostsByMetric: r.HostsByMetric,
		Hosts:         r.Hosts,
		UniqueMetrics: uniqMetricNames,
	})
}

// Host is the domain model for a host.
type Host struct {
	Name        string                `json:"name,omitempty"`
	MemcacheKey MemcacheKey           `json:"memcache_key,omitempty"`
	Metrics     map[MetricName]Metric `json:"metrics"`
	// Use a pointer for Mtime so that the json marshalling will omit when
	// empty.
	Mtime *time.Time `json:"mtime,omitempty"`
}

// Metric is the domain model for a metric.
type Metric struct {
	Name      string `json:"name,omitempty"`
	Datatype  string `json:"datatype,omitempty"`
	Units     string `json:"units,omitempty"`
	Source    string `json:"source,omitempty"`
	Value     string `json:"value,omitempty"`
	Nature    string `json:"nature,omitempty"`
	Dmax      int    `json:"dmax,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Stale     bool   `json:"stale"`
}

// MetricFromGanglia parses the given Ganglia metric into a format suitable
// for processing.
//
// The formats are largely similar with the following changes:
//
// Slope is replaced with a simplified Nature.
// TN is replaced with a more easily consumed Timestamp.
// TMAX is replaced with a more easily consumed Stale.
func MetricFromGanglia(now int64, src retrieval.Metric) (Metric, error) {
	var dst Metric
	tn, err := strconv.Atoi(src.TN)
	if err != nil {
		return Metric{}, errors.Wrap(err, "invalid TN")
	}
	tmax, err := strconv.Atoi(src.TMax)
	if err != nil {
		return Metric{}, errors.Wrap(err, "invalid TMAX")
	}
	dmax, err := strconv.Atoi(src.DMax)
	if err != nil {
		return Metric{}, errors.Wrap(err, "invalid DMAX")
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
