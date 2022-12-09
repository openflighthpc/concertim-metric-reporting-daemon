//go:generate go-enum --marshal --lower --names

package gds

import (
	"encoding/xml"
	"strconv"
	"time"
)

// MetricSlope describes how the value of the metric can change overtime.
//
// MetricSlopeZero: values are not comaprible, e.g., operating system name.
// MetricSlopePositive: value only increases over time, e.g., total downloads.
// MetricSlopeNegative: value only decreases over time.
// MetricSlopeBoth: value can increase or decrease over time.
// MetricSlopeDerivative: XXX What is this for?
//
// ENUM(zero, positive, negative, both, derivative).
type MetricSlope string

// MetricType describes the data type of the metric.
//
// ENUM(string, int8, uint8, int16, uint16, int32, uint32, float, double).
type MetricType string

type timestamp time.Time

// MarshalText implements the text marshaller method.
func (ts timestamp) MarshalText() ([]byte, error) {
	tmp := time.Time(ts)
	return []byte(strconv.FormatInt(tmp.Unix(), 10)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (ts *timestamp) UnmarshalText(text []byte) error {
	tmp, err := time.Parse(time.UnixDate, string(text))
	if err != nil {
		return err
	}
	*ts = timestamp(tmp)
	return nil
}

type cluster struct {
	XMLName   xml.Name  `xml:"CLUSTER"`
	Name      string    `xml:"NAME,attr"`
	LocalTime timestamp `xml:"LOCALTIME,attr"`
	Owner     string    `xml:"OWNER,attr"`
	LatLong   string    `xml:"LATLONG,attr"`
	URL       string    `xml:"URL,attr"`
	Hosts     []host    `xml:"HOSTS"`
}

type host struct {
	XMLName  xml.Name      `xml:"HOST"`
	Name     string        `xml:"NAME,attr"`
	IP       string        `xml:"IP,attr"`
	Reported timestamp     `xml:"REPORTED,attr"`
	Tn       time.Duration `xml:"TN,attr"`
	TMax     time.Duration `xml:"TMAX,attr"`
	DMax     time.Duration `xml:"DMAX,attr"`
	Metrics  []metric      `xml:"METRICS"`
}

type metric struct {
	XMLName xml.Name      `xml:"METRIC"`
	Name    string        `xml:"NAME,attr"`
	Val     string        `xml:"VAL,attr"`
	Units   string        `xml:"UNITS,attr"`
	Slope   MetricSlope   `xml:"SLOPE,attr"`
	Tn      time.Duration `xml:"TN,attr"`
	TMax    time.Duration `xml:"TMAX,attr"`
	DMax    time.Duration `xml:"DMAX,attr"`
	Source  string        `xml:"SOURCE,attr"`
	Type    MetricType    `xml:"TYPE,attr"`
}
