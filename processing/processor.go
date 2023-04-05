package main

import (
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/retrieval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Processor struct {
	dsmRepo *DSMRepo
	logger  zerolog.Logger
}

func NewProcessor(logger zerolog.Logger, dsmRepo *DSMRepo) *Processor {
	return &Processor{
		dsmRepo: dsmRepo,
		logger:  logger.With().Str("component", "processor").Logger(),
	}
}

func (p *Processor) Process(grids []retrieval.Grid) error {
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
			result.numIgnoredGrids += 1
			continue
		}
		p.logger.Debug().
			Str("grid", gGrid.Name).
			Int("count", len(gGrid.Clusters)).
			Msg("processing clusters")
		for _, gCluster := range gGrid.Clusters {
			if gCluster.Name != "unspecified" {
				p.logger.Warn().Str("cluster", gCluster.Name).Msg("ignoring")
				result.numIgnoredClusters += 1
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
					result.numIgnoredHosts += 1
					continue
				}
				p.logger.Debug().
					Str("grid", gGrid.Name).
					Str("cluster", gCluster.Name).
					Str("host", gHost.Name).
					Int("count", len(gHost.Metrics)).
					Msg("processing metrics")
				ctHost := Host{
					Name:    gHost.Name,
					Metrics: make(map[MetricName]Metric),
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
					ctHost.Mtime = time.Now()
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
	return nil
}

type (
	MetricName  string
	MemcacheKey string
)

// Result contains the result of the processing run for a single set of
// metrics.
type Result struct {
	// HostsByMetric is a map from a metric's name to a list of hosts that
	// currently have a fresh value for that metric.
	HostsByMetric map[MetricName][]MemcacheKey

	// UniqueMetrics is a set of unique metrics by name.
	UniqueMetrics map[MetricName]Metric

	// Hosts is a slice of Host containing their processed metrics.
	Hosts []Host

	// numMetrics keeps track of the total metrics processed for logging.  Other
	// logged information is derived from the other attributes.
	numMetrics         int
	numIgnoredGrids    int
	numIgnoredClusters int
	numIgnoredHosts    int
}

func NewResult() *Result {
	return &Result{
		HostsByMetric: map[MetricName][]MemcacheKey{},
		UniqueMetrics: map[MetricName]Metric{},
	}
}

func (r *Result) AddMetric(mckey MemcacheKey, metric Metric) {
	metricName := MetricName(metric.Name)
	hosts, ok := r.HostsByMetric[metricName]
	if !ok {
		hosts = make([]MemcacheKey, 0)
		r.HostsByMetric[metricName] = hosts
	}
	r.HostsByMetric[metricName] = append(hosts, mckey)
	r.UniqueMetrics[metricName] = metric
	r.numMetrics += 1
}

type Host struct {
	Name    string
	Metrics map[MetricName]Metric
	Mtime   time.Time
}

type Metric struct {
	Name      string
	Datatype  string
	Units     string
	Source    string
	Value     string
	Nature    string
	Dmax      int
	Timestamp int64
	Stale     bool
}

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
