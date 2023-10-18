package api

import (
	"errors"
	"math"
	"net/http"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

type historicHostResponse struct {
	Id     string                  `json:"id"`
	Values []historicValueResponse `json:"values"`
}

type historicValueResponse struct {
	Timestamp int64 `json:"timestamp"`
	Value     any   `json:"value"`
}

// getHistoricMetricValues returns a JSON list of historic metric values for
// all hosts that have reported the given metric between the given start and
// end times.
//
//	[
//	  {
//	    "id": "1",
//	    "values": [
//	      {
//	        "timestamp": 1696431225,
//	        "value": 9020
//	      },
//	      ...
//	    ]
//	  },
//	  ...
//	]
func (s *Server) getHistoricMetricValues(rw http.ResponseWriter, r *http.Request) {
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	startTime, err := parseTime(chi.URLParam(r, "startTime"))
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	endTime, err := parseTime(chi.URLParam(r, "endTime"))
	if err != nil {
		BadRequest(rw, r, err, "")
		return
	}
	duration := domain.HistoricMetricDurationFromTimes(startTime, endTime)
	s.fetchAndRenderMetrics(rw, r, metricName, duration)
}

func (s *Server) getHistoricMetricValuesLastX(rw http.ResponseWriter, r *http.Request) {
	metricName := domain.MetricName(chi.URLParam(r, "metricName"))
	lastX := chi.URLParam(r, "duration")
	duration, err := domain.HistoricMetricDurationFromString(lastX)
	if err != nil {
		if errors.Is(err, domain.ErrLastXLookupMissingEntry) {
			InternalError(rw, r, err)
		} else {
			BadRequest(rw, r, err, "")
		}
		return
	}
	s.fetchAndRenderMetrics(rw, r, metricName, duration)
}

func (s *Server) fetchAndRenderMetrics(
	rw http.ResponseWriter,
	r *http.Request,
	metricName domain.MetricName,
	duration domain.HistoricMetricDuration,
) {
	hosts, err := s.app.HistoricRepo.GetValuesForMetric(metricName, duration)
	if err != nil {
		InternalError(rw, r, err)
		return
	}
	body := []historicHostResponse{}
	for _, host := range hosts {
		body = append(body, historicHostResponseFromHistoricHost(host))
	}
	renderJSON(body, http.StatusOK, rw)
}

func historicHostResponseFromHistoricHost(src *domain.HistoricHost) historicHostResponse {
	var dst historicHostResponse
	dst.Id = src.Id.String()
	dst.Values = make([]historicValueResponse, 0, len(src.Metrics))

	for _, metrics := range src.Metrics {
		for _, metric := range metrics {
			dst.Values = append(dst.Values, historicValueResponseFromHistoricMetric(metric))
		}
	}
	return dst
}

func historicValueResponseFromHistoricMetric(src *domain.HistoricMetric) historicValueResponse {
	var dst historicValueResponse
	dst.Timestamp = src.Timestamp
	if math.IsNaN(src.Value) {
		dst.Value = nil
	} else {
		dst.Value = src.Value
	}
	return dst
}
