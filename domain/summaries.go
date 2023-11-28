package domain

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/rs/zerolog"
)

type metricSummaries struct {
	logger    zerolog.Logger
	summaries map[MetricName]*MetricSummary
}

var _ MetricSummaries = (*metricSummaries)(nil)

func newMetricSummaries() *metricSummaries {
	ms := metricSummaries{
		summaries: map[MetricName]*MetricSummary{},
	}
	return &ms
}

// AddMetric method implements the MetricSummaries interface.
func (ms *metricSummaries) AddMetric(metric CurrentMetric) error {
	ms.logger.Debug().Str("metric", metric.Name).Msg("adding metric")
	metricName := MetricName(metric.Name)
	summary, ok := ms.summaries[metricName]
	if !ok {
		summary = &MetricSummary{}
		ms.summaries[metricName] = summary
	}
	summary.Num += 1
	return addMetricValueToSum(summary, metric)
}

// AddMetric method implements the MetricSummaries interface.
func (ms *metricSummaries) GetSummaries() map[MetricName]*MetricSummary {
	return ms.summaries
}

// addMetricValueToSum adds the metrics Value to the summaries Sum attribute.
// Hoops are jumped through to handle the various data types.
func addMetricValueToSum(summary *MetricSummary, metric CurrentMetric) error {
	switch metric.Datatype {
	case "int8", "int16", "int32":
		i, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return err
		}
		if summary.Sum == nil {
			summary.Sum = i
		} else {
			sumVal := reflect.ValueOf(summary.Sum)
			if sumVal.CanInt() {
				summary.Sum = sumVal.Int() + i
			}
		}
		return nil
	case "uint8", "uint16", "uint32":
		i, err := strconv.ParseUint(metric.Value, 10, 64)
		if err != nil {
			return err
		}
		if summary.Sum == nil {
			summary.Sum = i
		} else {
			sumVal := reflect.ValueOf(summary.Sum)
			if sumVal.CanUint() {
				summary.Sum = sumVal.Uint() + i
			}
		}
		return nil
	case "float", "double":
		i, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return err
		}
		if summary.Sum == nil {
			summary.Sum = i
		} else {
			sumVal := reflect.ValueOf(summary.Sum)
			if sumVal.CanFloat() {
				summary.Sum = sumVal.Float() + i
			}
		}
		return nil
	default:
		return fmt.Errorf("unexpected metric type %s", metric.Datatype)
	}
}
