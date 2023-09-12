package api

import (
	"net/http"
)

type uniqueMetric struct {
	Id       string `json:"id"`
	Max      any    `json:"max"`
	Min      any    `json:"min"`
	Name     string `json:"name"`
	Nature   string `json:"nature"`
	Units    string `json:"units"`
}

// getUniqueMetrics returns a JSON list of unique metrics.  The uniqueness of
// metrics is determined by its metric name.  The format of the JSON is as follows:
//
//	[
//	  {
//	    "id": "caffeine.capacity",
//	    "name": "caffeine.capacity",
//	    "units": "",
//	    "nature": "volatile",
//	    "min": 32,
//	    "max": 64
//	  },
//	  ...
//	]
func (s *Server) getUniqueMetrics(rw http.ResponseWriter, r *http.Request) {
	metrics := s.app.ResultRepo.GetUniqueMetrics()
	body := []uniqueMetric{}
	for _, metric := range metrics {
		um := uniqueMetric{
			Id:       metric.Name,
			Name:     metric.Name,
			Nature:   metric.Nature,
			Units:    metric.Units,
			Min:      metric.Min,
			Max:      metric.Max,
		}
		body = append(body, um)
	}
	renderJSON(body, http.StatusOK, rw)
}
