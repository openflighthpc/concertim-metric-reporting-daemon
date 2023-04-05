// Package retrieval provides functions to periodically retrieve and parse
// ganglia XML.
package retrieval

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/net/html/charset"
)

type xmlRetriever interface {
	retrieve() ([]byte, error)
	describe() string
}

// Poller represents a structure that can periodically retrieve and parse
// ganglia XML.
//
// It has a single interesting method, Start, which starts the periodic
// retrieval of ganglia metrics.
//
// Its config contains the source of the ganglia data to retrieve and the
// period with which to retrieve it.
type Poller struct {
	config       config.Retrieval
	logger       zerolog.Logger
	xmlRetriever xmlRetriever
}

// New returns a new Poller.
func New(logger zerolog.Logger, config config.Retrieval) (*Poller, error) {
	xmlRetriever, err := getXMLRetriver(logger, config)
	if err != nil {
		return nil, err
	}

	return &Poller{
		config:       config,
		logger:       logger.With().Str("component", "metric-retriever").Logger(),
		xmlRetriever: xmlRetriever,
		// stopChan:  make(chan struct{}),
	}, nil
}

// Start periodically retrieves the ganglia XML, parses it and sends the
// results to the gridChan channel.
func (r *Poller) Start(gridChan chan<- []Grid) error {
	for {
		xml, err := r.xmlRetriever.retrieve()
		if err != nil {
			return err
		}
		grids, err := r.parseXML(xml)
		if err != nil {
			return err
		}
		r.logRetrieved(xml, grids)
		gridChan <- grids
		time.Sleep(r.config.Sleep)
	}
}

func (r *Poller) parseXML(gangliaXML []byte) ([]Grid, error) {
	r.logger.Debug().Int("bytes", len(gangliaXML)).Msg("parsing xml")
	var root gangliaRoot
	reader := bytes.NewReader(gangliaXML)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&root)
	if err != nil {
		return nil, errors.Wrap(err, "parsing ganglia xml")
	}
	return root.Grids, nil
}

func (r *Poller) logRetrieved(xml []byte, grids []Grid) {
	numHosts := 0
	numMetrics := 0
	for _, g := range grids {
		for _, c := range g.Clusters {
			for _, h := range c.Hosts {
				numHosts++
				for range h.Metrics {
					numMetrics++
				}
			}
		}
	}
	r.logger.Info().
		Int("bytes", len(xml)).
		Int("hosts", numHosts).
		Int("metrics", numMetrics).
		Str("source", r.xmlRetriever.describe()).
		Msg("completed")
}

func getXMLRetriver(logger zerolog.Logger, config config.Retrieval) (xmlRetriever, error) {
	if config.Testdata != "" {
		return &fileRetreiver{
			path:   config.Testdata,
			logger: logger,
		}, nil
	}
	return &tcpRetriever{
		addr:   fmt.Sprintf("%s:%d", config.IP, config.Port),
		logger: logger,
	}, nil
}
