// Package processing collects stuff about processing.
package processing

import (
	"slices"

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
// These views are stored in the given domain.ProcessedRepository.
type Processor struct {
	historicRepo domain.HistoricRepository
	logger       zerolog.Logger
	resultRepo   domain.ProcessedRepository
}

// NewProcessor returns a new *Processor.
func NewProcessor(
	resultRepo domain.ProcessedRepository,
	historicRepo domain.HistoricRepository,
	logger zerolog.Logger,
) *Processor {
	return &Processor{
		historicRepo: historicRepo,
		logger:       logger.With().Str("component", "processor").Logger(),
		resultRepo:   resultRepo,
	}
}

// Process the provided hosts to produce the expected views and store them in
// resultRepo.
func (p *Processor) Process(hosts []*domain.ProcessedHost) {
	stats := processLogStats{}
	summaries := newMetricSummaries()
	p.logger.Debug().Int("count", len(hosts)).Msg("processing hosts")
	err := p.resultRepo.Begin()
	if err != nil {
		p.logger.Error().Err(err).Msg("starting transaction")
		return
	}

	for _, host := range hosts {
		stats.numHosts++
		p.logger.Debug().
			Str("host", host.DSM.HostName).
			Int("count", len(host.Metrics)).
			Msg("processing metrics")
		for _, metric := range host.Metrics {
			stats.numMetrics++
			p.logger.Debug().
				Str("host", host.DSM.HostName).
				Str("metric", metric.Name).
				Msg("processing metric")
			p.resultRepo.AddMetric(host, &metric)
			if slices.Contains(domain.NumericMetricTypes, metric.Datatype) {
				if err := p.historicRepo.UpdateMetric(host, &metric); err != nil {
					p.logger.Warn().Err(err).Msg("updating historic repo")
				}
				if err = summaries.AddMetric(metric); err != nil {
					p.logger.Warn().Err(err).Msg("consolidating metric")
				}
			}
		}
		p.resultRepo.AddHost(host)
	}
	err = p.historicRepo.UpdateSummaryMetrics(summaries)
	if err != nil {
		p.logger.Error().Err(err).Msg("writing consolidated metrics")
	}
	err = p.resultRepo.Commit()
	if err != nil {
		p.logger.Error().Err(err).Msg("committing transaction")
		return
	}
	um, err := p.resultRepo.GetUniqueMetrics()
	if err == nil {
		stats.numUniqueMetrics = len(um)
	}
	logProcessResults(p.logger, stats)
}

func logProcessResults(logger zerolog.Logger, stats processLogStats) {
	logger.Info().
		Int("hosts", stats.numHosts).
		Int("metrics", stats.numMetrics).
		Int("unique metrics", stats.numUniqueMetrics).
		Msg("completed")
}

// processLogStats contains information useful for logging the results of the
// processing run.
type processLogStats struct {
	numHosts         int
	numMetrics       int
	numUniqueMetrics int
}
