package inmem

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

type processingResult struct {
	// hostsByMetric is a map from a metric's name to a list of hosts that
	// currently have a fresh value for that metric.
	hostsByMetric map[domain.MetricName][]*domain.ProcessedHost

	// uniqueMetrics is a set of unique metrics by name.
	uniqueMetrics map[domain.MetricName]*domain.UniqueMetric

	// hosts is a slice of Host containing their processed metrics.
	hosts []*domain.ProcessedHost
}

var _ domain.ProcessedRepository = (*ProcessedRepository)(nil)

type ProcessedRepository struct {
	logger     zerolog.Logger
	mux        sync.Mutex
	result     *processingResult
	nextResult *processingResult
}

func NewProcessedRepository(logger zerolog.Logger) *ProcessedRepository {
	return &ProcessedRepository{
		logger: logger.With().Str("component", "processed-repo").Logger(),
		mux:    sync.Mutex{},
	}
}

func (pr *ProcessedRepository) Begin() error {
	pr.logger.Debug().Msg("begin transaction")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.nextResult = &processingResult{
		hostsByMetric: map[domain.MetricName][]*domain.ProcessedHost{},
		uniqueMetrics: map[domain.MetricName]*domain.UniqueMetric{},
	}
	return nil
}

func (pr *ProcessedRepository) Commit() error {
	pr.logger.Debug().Any("results", pr.nextResult).Msg("committing transaction")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.result = pr.nextResult
	pr.nextResult = nil
	return nil
}

func (pr *ProcessedRepository) AddHost(host *domain.ProcessedHost) {
	// XXX Add error handling if outside of transaction.
	// if pr.nextResult == nil {
	// 	return fmt.Errorf("adding host outside of transaction")
	// }
	pr.logger.Debug().Stringer("host", host.DSM).Msg("adding host")
	pr.nextResult.hosts = append(pr.nextResult.hosts, host)
	if len(host.Metrics) > 0 {
		now := time.Now()
		host.Mtime = &now
	}
}

// AddMetric adds the given metric for the given host.  Host should be the host
// that reported the metric.
func (pr *ProcessedRepository) AddMetric(host *domain.ProcessedHost, metric *domain.ProcessedMetric) {
	// XXX Add error handling if outside of transaction.
	// if pr.nextResult == nil {
	// 	return fmt.Errorf("adding host outside of transaction")
	// }
	pr.logger.Debug().Stringer("host", host.DSM).Str("metric", metric.Name).Msg("adding metric")
	nextResult := pr.nextResult
	metricName := domain.MetricName(metric.Name)
	host.Metrics[metricName] = *metric
	hosts, ok := nextResult.hostsByMetric[metricName]
	if !ok {
		hosts = make([]*domain.ProcessedHost, 0)
		nextResult.hostsByMetric[metricName] = hosts
	}
	nextResult.hostsByMetric[metricName] = append(hosts, host)
	um, found := nextResult.uniqueMetrics[metricName]
	if !found {
		um = uniqueMetricFromMetric(*metric)
		pr.logger.Debug().Str("metric", um.Name).Msg("adding unique metric")
		nextResult.uniqueMetrics[metricName] = um
	}
	adjustMinMax(um, *metric)
	// pr.numMetrics++
}

func (pr *ProcessedRepository) GetUniqueMetrics() ([]*domain.UniqueMetric, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	metrics := make([]*domain.UniqueMetric, 0, len(pr.result.uniqueMetrics))
	for _, metric := range pr.result.uniqueMetrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (pr *ProcessedRepository) HostsWithMetric(metric domain.MetricName) ([]*domain.ProcessedHost, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	hosts, ok := pr.result.hostsByMetric[domain.MetricName(metric)]
	if !ok {
		return nil, domain.ErrMetricNotFound
	}
	return hosts, nil
}

func (pr *ProcessedRepository) GetMetricsForHost(hostId domain.HostId) ([]*domain.ProcessedMetric, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	var host *domain.ProcessedHost
	for _, candidate := range pr.result.hosts {
		if hostId == candidate.Id {
			host = candidate
			break
		}
	}
	if host == nil {
		return nil, domain.ErrHostNotFound
	}
	metrics := make([]*domain.ProcessedMetric, 0, len(host.Metrics))
	for _, metric := range host.Metrics {
		metric := metric
		metrics = append(metrics, &metric)
	}
	return metrics, nil
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
		// i, err := strconv.ParseInt(metric.Value, 10, 64)
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
	case "uint8", "uint16", "uint32":
		i, err := strconv.ParseUint(metric.Value, 10, 64)
		if err != nil {
			return
		}
		if unique.Min == nil {
			unique.Min = i
		} else {
			minVal := reflect.ValueOf(unique.Min)
			if minVal.CanUint() {
				if uint64(i) < minVal.Uint() {
					unique.Min = i
				}
			}
		}
		if unique.Max == nil {
			unique.Max = i
		} else {
			maxVal := reflect.ValueOf(unique.Max)
			if maxVal.CanUint() {
				if uint64(i) > maxVal.Uint() {
					unique.Max = i
				}
			}
		}
	case "float", "double":
		i, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return
		}
		if unique.Min == nil {
			unique.Min = i
		} else {
			minVal := reflect.ValueOf(unique.Min)
			if minVal.CanFloat() {
				if float64(i) < minVal.Float() {
					unique.Min = i
				}
			}
		}
		if unique.Max == nil {
			unique.Max = i
		} else {
			maxVal := reflect.ValueOf(unique.Max)
			if maxVal.CanFloat() {
				if float64(i) > maxVal.Float() {
					unique.Max = i
				}
			}
		}
	}
}
