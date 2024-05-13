//==============================================================================
// Copyright (C) 2024-present Alces Flight Ltd.
//
// This file is part of Concertim Metric Reporting Daemon.
//
// This program and the accompanying materials are made available under
// the terms of the Eclipse Public License 2.0 which is available at
// <https://www.eclipse.org/legal/epl-2.0>, or alternative license
// terms made available by Alces Flight Ltd - please direct inquiries
// about licensing to licensing@alces-flight.com.
//
// Concertim Metric Reporting Daemon is distributed in the hope that it will be useful, but
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
// IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
// OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
// PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
// details.
//
// You should have received a copy of the Eclipse Public License 2.0
// along with Concertim Metric Reporting Daemon. If not, see:
//
//  https://opensource.org/licenses/EPL-2.0
//
// For more information on Concertim Metric Reporting Daemon, please visit:
// https://github.com/openflighthpc/concertim-metric-reporting-daemon
//==============================================================================

package inmem

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/openflighthpc/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

// processingResult is a struct that holds the result of a single processing run.
type processingResult struct {
	// hostsByMetric is a map from a metric's name to a list of hosts that
	// currently have a fresh value for that metric.
	hostsByMetric map[domain.MetricName][]*domain.CurrentHost

	// uniqueMetrics is a set of unique metrics by name.
	uniqueMetrics map[domain.MetricName]*domain.UniqueMetric

	// hosts is a slice of CurrentHosts.  Each host contains its current metrics.
	hosts []*domain.CurrentHost
}

var _ domain.CurrentRepository = (*CurrentRepository)(nil)

// CurrentRepository implements the domain.CurrentRepository interface.  The
// results from the most recently completed processing run, if any, are stored
// in the result field.  The ongoing processing run, if any, is stored in the
// nextResult field.
type CurrentRepository struct {
	logger     zerolog.Logger
	mux        sync.Mutex
	result     *processingResult
	nextResult *processingResult
}

func NewCurrentRepository(logger zerolog.Logger) *CurrentRepository {
	return &CurrentRepository{
		logger: logger.With().Str("component", "current-repo").Logger(),
		mux:    sync.Mutex{},
	}
}

func (pr *CurrentRepository) Begin() error {
	pr.logger.Debug().Msg("begin transaction")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.nextResult = &processingResult{
		hostsByMetric: map[domain.MetricName][]*domain.CurrentHost{},
		uniqueMetrics: map[domain.MetricName]*domain.UniqueMetric{},
	}
	return nil
}

func (pr *CurrentRepository) Commit() error {
	pr.logger.Debug().Any("results", pr.nextResult).Msg("committing transaction")
	pr.mux.Lock()
	defer pr.mux.Unlock()
	pr.result = pr.nextResult
	pr.nextResult = nil
	return nil
}

func (pr *CurrentRepository) AddHost(host *domain.CurrentHost) {
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
func (pr *CurrentRepository) AddMetric(host *domain.CurrentHost, metric *domain.CurrentMetric) {
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
		hosts = make([]*domain.CurrentHost, 0)
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

func (pr *CurrentRepository) GetUniqueMetrics() ([]*domain.UniqueMetric, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	metrics := make([]*domain.UniqueMetric, 0, len(pr.result.uniqueMetrics))
	for _, metric := range pr.result.uniqueMetrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (pr *CurrentRepository) HostsWithMetric(metric domain.MetricName) ([]*domain.CurrentHost, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	hosts, ok := pr.result.hostsByMetric[domain.MetricName(metric)]
	if !ok {
		return nil, domain.ErrMetricNotFound
	}
	return hosts, nil
}

func (pr *CurrentRepository) GetMetricsForHost(hostId domain.HostId) ([]*domain.CurrentMetric, error) {
	if pr.result == nil {
		return nil, domain.ErrWaitingOnProcessingRun
	}
	var host *domain.CurrentHost
	for _, candidate := range pr.result.hosts {
		if hostId == candidate.Id {
			host = candidate
			break
		}
	}
	if host == nil {
		return nil, domain.ErrHostNotFound
	}
	metrics := make([]*domain.CurrentMetric, 0, len(host.Metrics))
	for _, metric := range host.Metrics {
		metric := metric
		metrics = append(metrics, &metric)
	}
	return metrics, nil
}

func uniqueMetricFromMetric(src domain.CurrentMetric) *domain.UniqueMetric {
	var dst domain.UniqueMetric
	dst.Datatype = src.Datatype
	dst.Name = src.Name
	dst.Nature = src.Nature
	dst.Units = src.Units
	adjustMinMax(&dst, src)
	return &dst
}

func adjustMinMax(unique *domain.UniqueMetric, metric domain.CurrentMetric) {
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
