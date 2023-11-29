package api

import (
	"errors"
	"net/http"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

// getCurrentHostMetrics returns a JSON list of all current metrics for the
// given host.
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity",
//	    "units": "cups",
//	    "nature": "volatile",
//	    "value": 3,
//	  },
//	  ...
//	]
func (s *Server) getCurrentHostMetrics(rw http.ResponseWriter, r *http.Request) {
	type metricResponse struct {
		Id     string `json:"id"`
		Name   string `json:"name"`
		Nature string `json:"nature"`
		Units  string `json:"units"`
		Value  any    `json:"value"`
	}
	body := []metricResponse{}
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
	metrics, err := s.app.CurrentRepo.GetMetricsForHost(hostId)
	if err != nil {
		if errors.Is(err, domain.ErrWaitingOnProcessingRun) {
			ServiceUnavailable(rw, r, err)
		} else if errors.Is(err, domain.ErrHostNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	for _, metric := range metrics {
		castValue, err := castMetricValue(*metric)
		if err != nil {
			s.logger.Debug().Err(err).Str("type", metric.Datatype).Str("value", metric.Value).Msg("error casting metric value")
			castValue = metric.Value
		}
		mr := metricResponse{
			Id:     metric.Name,
			Name:   metric.Name,
			Nature: metric.Nature,
			Units:  metric.Units,
			Value:  castValue,
		}
		body = append(body, mr)
	}
	renderJSON(body, http.StatusOK, rw)
}
