package processing

import (
	"sync"

	"github.com/rs/zerolog"
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
	mr.logger.Debug().Msg("recording results")
	mr.mux.Lock()
	defer mr.mux.Unlock()
	mr.result = result
	return nil
}
