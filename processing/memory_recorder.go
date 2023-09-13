package processing

import (
	"sync"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

var (
	_ Recorder          = (*MemoryRecorder)(nil)
	_ domain.ResultRepo = (*MemoryRecorder)(nil)
)

type MemoryRecorder struct {
	logger zerolog.Logger
	mux    sync.Mutex
	result *Result
}

func NewMemoryRecorder(logger zerolog.Logger) *MemoryRecorder {
	return &MemoryRecorder{
		logger: logger.With().Str("component", "memory.recorder").Logger(),
		mux:    sync.Mutex{},
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
	// XXX Better error handling.
	if mr.result == nil {
		// XXX NotReady response here.
		return nil
	}
	metrics := make([]domain.UniqueMetric, 0, len(mr.result.UniqueMetrics))
	for _, metric := range mr.result.UniqueMetrics {
		metrics = append(metrics, *metric)
	}
	return metrics
}

func (mr *MemoryRecorder) HostsWithMetric(metric domain.MetricName) []*domain.ProcessedHost {
	// XXX Better error handling.
	if mr.result == nil {
		// XXX NotReady response here.
		return nil
	}
	hosts, ok := mr.result.HostsByMetric[domain.MetricName(metric)]
	if !ok {
	// XXX 404 here.
		return make([]*domain.ProcessedHost, 0)
	}
	return hosts
}
