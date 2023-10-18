package api

import (
	"errors"
	"net/http"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

// getHistoricHostMetricValues returns a JSON list of historic metric values
// for the given host and metric between the given start and end times.
//
// [
//
//	  {
//	    "timestamp": 1696431225,
//	    "value": 9020
//		 },
//		 ...
//
// ]
func (s *Server) getHistoricHostMetricValues(rw http.ResponseWriter, r *http.Request) {
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
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
	s.fetchAndRenderHostMetrics(rw, r, hostId, metricName, duration)
}

func (s *Server) getHistoricHostMetricValuesLastX(rw http.ResponseWriter, r *http.Request) {
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
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
	s.fetchAndRenderHostMetrics(rw, r, hostId, metricName, duration)
}

func (s *Server) fetchAndRenderHostMetrics(
	rw http.ResponseWriter,
	r *http.Request,
	hostId domain.HostId,
	metricName domain.MetricName,
	duration domain.HistoricMetricDuration,
) {
	host, err := s.app.HistoricRepo.GetValuesForHostAndMetric(hostId, metricName, duration)
	if err != nil {
		if errors.Is(err, domain.ErrHostNotFound) {
			NotFound(rw, r, err)
		} else if errors.Is(err, domain.ErrMetricNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	body := []historicValueResponse{}
	metrics, ok := host.Metrics[metricName]
	if !ok {
		NotFound(rw, r, domain.ErrMetricNotFound)
	}
	for _, metric := range metrics {
		body = append(body, historicValueResponseFromHistoricMetric(metric))
	}
	renderJSON(body, http.StatusOK, rw)
}