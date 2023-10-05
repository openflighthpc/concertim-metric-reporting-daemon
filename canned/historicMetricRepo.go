package canned

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
)

type historicMetricRepo struct {
}

func NewHistoricMetricRepo() *historicMetricRepo {
	return &historicMetricRepo{}
}

func (historicMetricRepo) GetValuesForMetric(metricName domain.MetricName, startTime, endTime time.Time) ([]*domain.HistoricHost, error) {
	hosts := make([]*domain.HistoricHost, 0)
	for i := 0; i < 3; i++ {
		host := domain.HistoricHost{
			Id: domain.HostId(strconv.Itoa(i)),
			DSM: domain.DSM{
				GridName:    "unspecified",
				ClusterName: "unspecified",
				HostName:    fmt.Sprintf("device:%d", i),
			},
			Metrics: map[domain.MetricName][]*domain.HistoricMetric{},
		}
		for j := 0; j < 2; j++ {
			hm := domain.HistoricMetric{Value: float64(j), Timestamp: time.Now().Unix()}
			host.Metrics[domain.MetricName(fmt.Sprintf("fake.metric.%d", j))] = make([]*domain.HistoricMetric, 1)
			host.Metrics[domain.MetricName(fmt.Sprintf("fake.metric.%d", j))][0] = &hm
		}
		hosts = append(hosts, &host)
	}
	fmt.Printf("hosts: %v\n", hosts)
	return hosts, nil
}
