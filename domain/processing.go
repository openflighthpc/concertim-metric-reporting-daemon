package domain

import (
	"slices"
	"time"

	"github.com/rs/zerolog"
)

// Processor is a struct for processing the pending metrics.
//
// It produces three different views of the metrics, which are stored in the
// given currentRepo.
//
//  1. A list of unique metrics.
//  2. For each unique metric, a list of which devices are currently reporting
//     that metric.
//  3. For each host a map from metric name to that metric.
//
// It updates the historicRepo with the value of any non-expired pending metrics.
//
// It also creates summary view of the metrics and updates the historicRepo
// with those summaries.
type Processor struct {
	currentRepo  CurrentRepository
	historicRepo HistoricRepository
	logger       zerolog.Logger
	pendingRepo  PendingRepository
	step         time.Duration
}

// NewProcessor returns a new *Processor.
func NewProcessor(
	pendingRepo PendingRepository,
	currentRepo CurrentRepository,
	historicRepo HistoricRepository,
	step time.Duration,
	logger zerolog.Logger,
) *Processor {
	return &Processor{
		currentRepo:  currentRepo,
		historicRepo: historicRepo,
		logger:       logger.With().Str("component", "processor").Logger(),
		pendingRepo:  pendingRepo,
		step:         step,
	}
}

// Process the pending repo.
func (p *Processor) Process() {
	start := time.Now()
	stats := processLogStats{}
	summaries := newMetricSummaries()
	pendingHosts := p.pendingRepo.GetAll()
	p.logger.Debug().Int("count", len(pendingHosts)).Msg("processing hosts")
	err := p.currentRepo.Begin()
	if err != nil {
		p.logger.Error().Err(err).Msg("starting transaction")
		return
	}

	for _, pendingHost := range pendingHosts {
		stats.numHosts++
		p.logger.Debug().
			Str("host", pendingHost.DSM.HostName).
			Int("count", len(pendingHost.Metrics)).
			Msg("processing metrics")
		host := processedHostFromPendingHost(pendingHost)
		for _, pendingMetric := range pendingHost.Metrics {
			stats.numMetrics++
			p.logger.Debug().
				Str("host", host.DSM.HostName).
				Str("metric", pendingMetric.Name).
				Msg("processing metric")

			metric := p.processedMetricFromPendingMetric(pendingMetric, start)
			p.logger.Debug().Any("pending", pendingMetric).Any("processed", metric).Send()
			if metric.Stale {
				p.logger.Debug().
					Str("host", host.DSM.HostName).
					Str("metric", metric.Name).
					Msg("stale metric")
				stats.numStaleMetrics++
				continue
			}
			err := p.pendingRepo.UpdateLastProcessed(pendingHost.Id, MetricName(pendingMetric.Name), metric.Timestamp)
			if err != nil {
				p.logger.Warn().Err(err).Msg("updating metric last processed timestamp")
			}
			p.currentRepo.AddMetric(&host, &metric)
			if slices.Contains(NumericMetricTypes, metric.Datatype) {
				if err := p.historicRepo.UpdateMetric(&host, &metric); err != nil {
					p.logger.Warn().Err(err).Msg("updating historic repo")
				}
				if err = summaries.AddMetric(metric); err != nil {
					p.logger.Warn().Err(err).Msg("consolidating metric")
				}
			}
		}
		p.currentRepo.AddHost(&host)
	}
	err = p.historicRepo.UpdateSummaryMetrics(summaries)
	if err != nil {
		p.logger.Error().Err(err).Msg("writing consolidated metrics")
	}
	err = p.currentRepo.Commit()
	if err != nil {
		p.logger.Error().Err(err).Msg("committing transaction")
		return
	}
	um, err := p.currentRepo.GetUniqueMetrics()
	if err == nil {
		stats.numUniqueMetrics = len(um)
	}
	logProcessResults(p.logger, stats, time.Since(start))
}

func logProcessResults(logger zerolog.Logger, stats processLogStats, duration time.Duration) {
	logger.Info().
		Dur("duration", duration).
		Int("hosts", stats.numHosts).
		Int("metrics", stats.numMetrics).
		Int("unique metrics", stats.numUniqueMetrics).
		Int("stale metrics", stats.numStaleMetrics).
		Msg("completed")
}

// processLogStats contains information useful for logging the results of the
// processing run.
type processLogStats struct {
	numHosts         int
	numMetrics       int
	numStaleMetrics  int
	numUniqueMetrics int
}

func (p *Processor) processedMetricFromPendingMetric(src PendingMetric, now time.Time) ProcessedMetric {
	var dst ProcessedMetric
	var stale bool
	expirationTime := src.Reported.Add(src.TTL)
	persistent := src.TTL == 0
	if persistent {
		stale = false
	} else {
		stale = expirationTime.Before(now)
	}
	nature := "volatile"
	if src.Type == "string" || src.Type == "timestamp" {
		nature = "string_and_time"
	} else if src.Slope == "zero" {
		nature = "constant"
	}
	var timestamp time.Time
	if stale {
		// We're not processing the metric as it is stale, so we don't update
		// its LastProcessed.
	} else if src.LastProcessed == nil {
		// This is the first time we're processing this metric.  It's timestamp
		// is whenever it was reported.
		timestamp = src.Reported
	} else {
		// This metric value has previously been reported and processed but has
		// not yet expired.  We will continue to process its last reported
		// value until either it expires or a new value is reported for it.
		//
		// Due to the way that RRD works, dst.Timestamp should be set to a
		// value that is both (1) src.Reported + an integer multiple of step;
		// and (2) not after the current time.  The time comparisons are done
		// on whole seconds as that is how RRDTool works.
		lastProcessed := *src.LastProcessed
		for !lastProcessed.Add(p.step).Truncate(1 * time.Second).After(now.Truncate(1 * time.Second)) {
			lastProcessed = lastProcessed.Add(p.step)
		}
		timestamp = lastProcessed
		p.logger.Debug().
			Str("metric", src.Name).
			Int64("processing start", now.Unix()).
			Int64("timestamp", timestamp.Unix()).
			Int64("reported", src.Reported.Unix()).
			Int64("expiration", expirationTime.Unix()).
			Msg("reprocessing non-expired metric")
	}

	dst.Name = src.Name
	dst.Datatype = src.Type.String()
	dst.Units = src.Units
	dst.Value = src.Value
	dst.Nature = nature
	dst.Timestamp = timestamp
	dst.Stale = stale
	return dst
}

func processedHostFromPendingHost(src PendingHost) ProcessedHost {
	var dst ProcessedHost
	dst.Id = src.Id
	dst.DSM = src.DSM
	dst.Metrics = map[MetricName]ProcessedMetric{}
	return dst
}
