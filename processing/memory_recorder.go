package processing

import (
	"sync"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/dsmRepository"
	"github.com/rs/zerolog"
)

var (
	_ Recorder          = (*MemoryRecorder)(nil)
	_ domain.ResultRepo = (*MemoryRecorder)(nil)
)

type MemoryRecorder struct {
	dsmRepo *dsmRepository.Repo
	logger  zerolog.Logger
	mux     sync.Mutex
	result  *Result
}

func NewMemoryRecorder(logger zerolog.Logger, dsmRepo *dsmRepository.Repo) *MemoryRecorder {
	return &MemoryRecorder{
		dsmRepo: dsmRepo,
		logger:  logger.With().Str("component", "memory.recorder").Logger(),
		mux:     sync.Mutex{},
	}
}

func (mr *MemoryRecorder) Record(result *Result) error {
	mr.logger.Debug().Any("results", result).Msg("recording results")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.result = result
	return nil
}

func (mr *MemoryRecorder) GetUniqueMetrics() []domain.UniqueMetric {
	if mr.result == nil {
		return nil
	}
	metrics := make([]domain.UniqueMetric, 0, len(mr.result.UniqueMetrics))
	for _, metric := range mr.result.UniqueMetrics {
		metrics = append(metrics, *metric)
	}
	return metrics
}
