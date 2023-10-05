// Package processing collects stuff about processing.
package processing

import (
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
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
// These views are stored in the given domain.ProcessedRepository.
type Processor struct {
	resultRepo domain.ProcessedRepository
	logger     zerolog.Logger
}

// NewProcessor returns a new *Processor.
func NewProcessor(resultRepo domain.ProcessedRepository, logger zerolog.Logger) *Processor {
	return &Processor{
		resultRepo: resultRepo,
		logger:     logger.With().Str("component", "processor").Logger(),
	}
}

// Process the provided hosts to produce the expected views and store them in
// resultRepo.
func (p *Processor) Process(hosts []*domain.ProcessedHost) error {
	stats := processStats{}
	p.logger.Debug().Int("count", len(hosts)).Msg("processing hosts")
	err := p.resultRepo.Begin()
	if err != nil {
		return errors.Wrap(err, "starting transaction")
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
		}
		p.resultRepo.AddHost(host)
	}
	err = p.resultRepo.Commit()
	if err != nil {
		return errors.Wrap(err, "committing transaction")
	}
	um, err := p.resultRepo.GetUniqueMetrics()
	if err != nil {
		stats.numUniqueMetrics = len(um)
	}
	logProcessResults(p.logger, stats)
	return nil
}

func logProcessResults(logger zerolog.Logger, stats processStats) {
	logger.Info().
		Int("hosts", stats.numHosts).
		Int("metrics", stats.numMetrics).
		Int("unique metrics", stats.numUniqueMetrics).
		Msg("completed")
}

type processStats struct {
	numHosts         int
	numMetrics       int
	numUniqueMetrics int
}
