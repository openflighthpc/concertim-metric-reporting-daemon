// Package retrieval provides functions to periodically retrieve and parse
// ganglia XML.
package retrieval

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/ticker"
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
	Ticker       *ticker.Ticker
	xmlRetriever xmlRetriever
}

// New returns a new Poller.
func New(logger zerolog.Logger, config config.Retrieval) (*Poller, error) {
	logger = logger.With().Str("component", "metric-retriever").Logger()
	xmlRetriever, err := getXMLRetriver(logger, config)
	if err != nil {
		return nil, errors.Wrap(err, "getting xml retriever")
	}

	return &Poller{
		config:       config,
		logger:       logger,
		Ticker:       ticker.NewTicker(config.Frequency, config.Throttle),
		xmlRetriever: xmlRetriever,
	}, nil
}

// Start periodically retrieves the ganglia XML, parses it and sends the
// results to the hostsChan channel.
func (r *Poller) Start(hostsChan chan<- []Host) {
	oneLoop := func() {
		xml, err := r.xmlRetriever.retrieve()
		if err != nil {
			r.logger.Err(err).Send()
			return
		}
		grids, err := r.parseXML(xml)
		if err != nil {
			r.logger.Err(err).Send()
			return
		}
		r.logRetrieved(xml, grids)
		hosts := r.extractHosts(grids)
		hostsChan <- hosts
	}
	for {
		<-r.Ticker.C
		oneLoop()
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

func (r *Poller) extractHosts(grids []Grid) []Host {
	r.logger.Debug().Int("count", len(grids)).Msg("filtering grids")
	hosts := make([]Host, 0)
	for _, grid := range grids {
		if grid.Name != r.config.GridName {
			r.logger.Warn().Str("grid", grid.Name).Msg("ignoring")
			continue
		}
		r.logger.Debug().
			Str("grid", grid.Name).
			Int("count", len(grid.Clusters)).
			Msg("filtering clusters")
		for _, cluster := range grid.Clusters {
			if cluster.Name != r.config.ClusterName {
				r.logger.Warn().Str("cluster", cluster.Name).Msg("ignoring")
				continue
			}
			hosts = append(hosts, cluster.Hosts...)
		}
	}

	return hosts
}
