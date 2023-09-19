package processing

import "github.com/alces-flight/concertim-metric-reporting-daemon/domain"

// Recorder is an interface for recording the processed metrics.
type Recorder interface {
	Record(*Result) error
}

type ResultRepo interface {
	AddHost(host *domain.ProcessedHost)
	AddMetric(host *domain.ProcessedHost, metric *domain.ProcessedMetric)
}
