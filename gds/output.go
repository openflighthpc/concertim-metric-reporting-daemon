package gds

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"text/template"
	"time"

	"github.com/alces-flight/concertim-mrapi/domain"
)

//go:embed output.tmpl
var outputTemplateBytes []byte

// clock is an interface describing the generator's dependencies on time.Time.
//
// It exists to effectively allow stubbing clock.Now() with a canned response
// in tests.
type clock interface {
	Now() time.Time
}

// realClock is used to implemented clock for time.Time.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type outputGenerator struct {
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

func newOutputGenerator(clock clock) (*outputGenerator, error) {
	funcMap := template.FuncMap{
		"localTime":    clock.Now().Unix,
		"timeToUnix":   func(t time.Time) int64 { return t.Unix() },
		"secondsAsInt": func(d time.Duration) int { return int(d.Seconds()) },
		"xml":          escapeXML,
	}
	outputTemplate, err := template.New("output").
		Funcs(funcMap).
		Parse(string(outputTemplateBytes))
	if err != nil {
		return nil, err
	}
	return &outputGenerator{template: outputTemplate}, nil
}

func (g *outputGenerator) generate(dCluster domain.Cluster) ([]byte, error) {
	var buf bytes.Buffer
	err := g.template.Execute(&buf, dCluster)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
