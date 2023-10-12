package api

import "net/http"

// getHistoricMetricNames returns a JSON list of historic metric names. The
// format of the JSON is as follows:
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity"
//	  },
//	  ...
//	]
func (s *Server) getHistoricMetricNames(rw http.ResponseWriter, r *http.Request) {
	type historicMetricName struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	metricNames, err := s.app.HistoricRepo.ListMetricNames()
	if err != nil {
		InternalError(rw, r, err)
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
