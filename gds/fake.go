package gds

import "time"

func newCluster(hosts []host) cluster {
	return cluster{
		Name:      "unspecified",
		LocalTime: timestamp(time.Now()),
		Owner:     "unspecified",
		LatLong:   "unspecified",
		URL:       "unspecified",
		Hosts:     hosts,
	}
}

func newHost(name string, tmax, dmax time.Duration, metrics []metric) host {
	return host{
		Name:     name,
		IP:       "",
		Reported: timestamp(time.Now()),
		Tn:       10,
		TMax:     tmax,
		DMax:     tmax,
		Metrics:  metrics,
	}
}

func fakeCluster() cluster {
	hosts := []host{
		newHost("comp001", 60, 60, []metric{
			metric{
				Name:   "foo",
				Val:    "foobar",
				Units:  "foos",
				Slope:  MetricSlopeZero,
				Tn:     0,
				TMax:   60,
				DMax:   60,
				Source: "MRAPI",
				Type:   MetricTypeString,
			},
			metric{
				Name:   "bar",
				Val:    "27",
				Units:  "carrots",
				Slope:  MetricSlopeZero,
				Tn:     10,
				TMax:   60,
				DMax:   60,
				Source: "MRAPI",
				Type:   MetricTypeString,
			},
		}),
		newHost("comp002", 60, 60, []metric{
			metric{
				Name:   "baz",
				Val:    "27",
				Units:  "turnips",
				Slope:  MetricSlopeZero,
				Tn:     10,
				TMax:   60,
				DMax:   60,
				Source: "MRAPI",
				Type:   MetricTypeString,
			},
		}),
	}
	return newCluster(hosts)
}
