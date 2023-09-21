package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

type metricValue struct {
	Id    string `json:"id"`
	Value any    `json:"value"`
}

// getMetricValues returns a JSON list of current values for metric.
//
//	[
//	  {
//	    "id": "42",
//	    "value": 32
//	  },
//	  ...
//	]
func (s *Server) getMetricValues(rw http.ResponseWriter, r *http.Request) {
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	hosts := s.app.ResultRepo.HostsWithMetric(metricName)
	if hosts == nil {
		body := ErrorsPayload{
			Status: http.StatusNotFound,
			Errors: []*ErrorObject{{Title: "Metric Not Found", Detail: "Metric Not Found"}},
		}
		renderJSON(body, http.StatusNotFound, rw)
		return
	}
	body := []metricValue{}
	for _, host := range hosts {
		metric, ok := host.Metrics[metricName]
		if !ok {
			s.logger.Warn().
				Stringer("host", host.Id).
				Str("metric", string(metricName)).
				Msg("metric not found for host")
			continue
		}
		castValue, err := castMetricValue(metric)
		if err != nil {
			s.logger.Debug().
				Err(err).
				Str("type", metric.Datatype).
				Str("value", metric.Value).
				Msg("error casting metric value")
			castValue = metric.Value
		}
		mv := metricValue{
			Id:    host.Id.String(),
			Value: castValue,
		}
		body = append(body, mv)
	}
	renderJSON(body, http.StatusOK, rw)
}

func castMetricValue(metric domain.ProcessedMetric) (any, error) {
	switch metric.Datatype {
	case "int8", "int16", "int32":
		i, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case "uint8", "uint16", "uint32":
		i, err := strconv.ParseUint(metric.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case "string":
		return metric.Value, nil
	case "float", "double":
		i, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	default:
		return nil, fmt.Errorf("unexpected metric type %s", metric.Datatype)
	}
}
