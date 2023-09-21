package gds

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"text/template"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
)

//go:embed output.tmpl
var outputTemplateBytes []byte

// clock is an interface describing the generator's dependencies on time.Time.
//
// It exists to effectively allow stubbing clock.Now() with a canned response
// in tests.
type clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// realClock is used to implemented clock for time.Time.
type realClock struct{}

func (realClock) Now() time.Time                  { return time.Now() }
func (realClock) Since(t time.Time) time.Duration { return time.Since(t) }

type outputGenerator struct {
	clock    clock
	config   config.GDS
	template *template.Template
}

func escapeXML(a any) any {
	var out bytes.Buffer
	switch aa := a.(type) {
	case []byte:
		err := xml.EscapeText(&out, aa)
		if err != nil {
			return ""
		}
		return out.String()
	case string:
		in := []byte(aa)
		err := xml.EscapeText(&out, in)
		if err != nil {
			return ""
		}
		return out.String()
	default:
		return a
	}
}

// secondsAsInt returns the duration as an integer number of seconds.
func secondsAsInt(d time.Duration) int {
	return int(d.Seconds())
}

func newOutputGenerator(clock clock, config config.GDS) (*outputGenerator, error) {
	og := outputGenerator{clock: clock, config: config}
	funcMap := template.FuncMap{
		"timeToUnix":   func(t time.Time) int64 { return t.Unix() },
		"secondsAsInt": secondsAsInt,
		"xml":          escapeXML,
		"secondsSince": og.secondsSince,
	}
	outputTemplate, err := template.New("output").
		Funcs(funcMap).
		Parse(string(outputTemplateBytes))
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}
	og.template = outputTemplate
	return &og, nil
}

// secondsSince returns an integer number of seconds since startTime.
func (g *outputGenerator) secondsSince(startTime time.Time) int {
	return secondsAsInt(g.clock.Since(startTime))
}

func (g *outputGenerator) generate(hosts []domain.ReportedHost) ([]byte, error) {
	var buf bytes.Buffer
	cluster := cluster{
		Hosts:        hosts,
		LocalTime:    g.clock.Now().Unix(),
		MetricSource: g.config.MetricSource,
		Name:         g.config.ClusterName,
	}
	err := g.template.Execute(&buf, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "executing template")
	}
	return buf.Bytes(), nil
}

type cluster struct {
	Hosts        []domain.ReportedHost
	LocalTime    int64
	MetricSource string
	Name         string
}
