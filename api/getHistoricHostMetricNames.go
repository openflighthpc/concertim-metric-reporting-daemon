package api

import (
	"errors"
	"net/http"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/go-chi/chi/v5"
)

// getHistoricHostMetricNames returns a JSON list of historic metric names. The
// format of the JSON is as follows:
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity"
//	  },
//	  ...
//	]
func (s *Server) getHistoricHostMetricNames(rw http.ResponseWriter, r *http.Request) {
	type historicMetricName struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	hostId := domain.HostId(chi.URLParam(r, "deviceId"))
	metricNames, err := s.app.HistoricRepo.ListHostMetricNames(hostId)
	if err != nil {
		if errors.Is(err, domain.ErrHostNotFound) {
			NotFound(rw, r, err)
		} else {
			InternalError(rw, r, err)
		}
		return
	}
	body := []historicMetricName{}
	for _, metricName := range metricNames {
		um := historicMetricName{
			Id:   metricName,
			Name: metricName,
		}
		body = append(body, um)
	}
	renderJSON(body, http.StatusOK, rw)
}
