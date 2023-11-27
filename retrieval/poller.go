// Package retrieval provides functions to periodically retrieve and parse
// ganglia XML.
package retrieval

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/net/html/charset"
	"golang.org/x/time/rate"
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
	Ticker       *time.Ticker
	config       config.Retrieval
	dsmRepo      domain.DataSourceMapRepository
	dsmUpdater   domain.DataSourceMapRepoUpdater
	hostsChan    chan<- []*domain.ProcessedHost
	limiter      rate.Sometimes
	logger       zerolog.Logger
	xmlRetriever xmlRetriever
}

// New returns a new Poller.
func New(
	logger zerolog.Logger,
	config config.Retrieval,
	dsmRepo domain.DataSourceMapRepository,
	dsmUpdater domain.DataSourceMapRepoUpdater,
	hostsChan chan<- []*domain.ProcessedHost,
) (*Poller, error) {
	logger = logger.With().Str("component", "metric-retriever").Logger()
	xmlRetriever, err := getXMLRetriver(logger, config)
	if err != nil {
		return nil, errors.Wrap(err, "getting xml retriever")
	}

	return &Poller{
		Ticker:       time.NewTicker(config.Frequency),
		config:       config,
		dsmRepo:      dsmRepo,
		dsmUpdater:   dsmUpdater,
		hostsChan:    hostsChan,
		limiter:      rate.Sometimes{First: 1, Interval: config.Throttle},
		logger:       logger,
		xmlRetriever: xmlRetriever,
	}, nil
}

// Start periodically retrieves the ganglia XML, parses it and sends the
// results to the hostsChan channel.
func (p *Poller) Start() {
	for {
		<-p.Ticker.C
		p.PollOnce()
	}
}

func (p *Poller) PollOnce() {
	p.limiter.Do(func() {
		p.dsmUpdater.UpdateNow()
		xml, err := p.xmlRetriever.retrieve()
		if err != nil {
			p.logger.Err(err).Send()
			return
		}
		grids, err := p.parseXML(xml)
		if err != nil {
			p.logger.Err(err).Send()
			return
		}
		p.logRetrieved(xml, grids)
		hosts := p.extractHosts(grids)
		p.hostsChan <- hosts
	})
}

func (p *Poller) parseXML(gangliaXML []byte) ([]Grid, error) {
	p.logger.Debug().Int("bytes", len(gangliaXML)).Msg("parsing xml")
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

func (p *Poller) logRetrieved(xml []byte, grids []Grid) {
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
	p.logger.Info().
		Int("bytes", len(xml)).
		Int("hosts", numHosts).
		Int("metrics", numMetrics).
		Str("source", p.xmlRetriever.describe()).
		Msg("retrieved")
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

type extractStats struct {
	parsedGrids     int
	parsedClusters  int
	parsedHosts     int
	parsedMetrics   int
	ignoredGrids    int
	ignoredClusters int
	failedHosts     int
	failedMetrics   int
}

// extractHosts returns a list of `domain.ProcessedHost`s extracted from the
// ganglia output.
//
// Only the hosts in a single grid and cluster are extracted.  That grid and
// cluster are specified in the configuration.
func (p *Poller) extractHosts(grids []Grid) []*domain.ProcessedHost {
	now := time.Now()
	stats := extractStats{}
	p.logger.Debug().Int("count", len(grids)).Msg("parsing grids")
	hosts := make([]*domain.ProcessedHost, 0)
	for _, grid := range grids {
		if grid.Name != p.config.GridName {
			p.logger.Warn().Str("grid", grid.Name).Msg("ignoring")
			stats.ignoredGrids++
			continue
		}
		stats.parsedGrids++
		p.logger.Debug().
			Str("grid", grid.Name).
			Int("count", len(grid.Clusters)).
			Msg("parsing clusters")
		for _, cluster := range grid.Clusters {
			if cluster.Name != p.config.ClusterName {
				p.logger.Warn().Str("cluster", cluster.Name).Msg("ignoring")
				stats.ignoredClusters++
				continue
			}
			stats.parsedClusters++
			for _, host := range cluster.Hosts {
				p.logger.Debug().Str("host", host.Name).Msg("parsing host")
				pHost, err := p.processedHostFromGanglia(grid, cluster, host, now)
				if err != nil {
					p.logger.Warn().Err(err).Str("host", host.Name).Msg("ignoring")
					stats.failedHosts++
					continue
				}
				stats.parsedHosts++
				for _, gMetric := range host.Metrics {
					p.logger.Debug().
						Str("host", host.Name).
						Str("metric", gMetric.Name).
						Msg("parsing metric")
					pMetric, err := processedMetricFromGanglia(now, gMetric)
					if err != nil {
						p.logger.Warn().Err(err).Msg("failed")
						stats.failedMetrics++
						continue
					}
					stats.parsedMetrics++
					pHost.Metrics[domain.MetricName(gMetric.Name)] = pMetric
				}
				hosts = append(hosts, pHost)
			}
		}
	}

	logStats(p.logger, stats)

	return hosts
}

func logStats(logger zerolog.Logger, stats extractStats) {
	logger.Info().
		Dict("parsed", zerolog.Dict().
			Int("grids", stats.parsedGrids).
			Int("clusters", stats.parsedClusters).
			Int("hosts", stats.parsedHosts).
			Int("metrics", stats.parsedMetrics)).
		Dict("ignored", zerolog.Dict().
			Int("grids", stats.ignoredGrids).
			Int("clusters", stats.ignoredClusters)).
		Dict("failed", zerolog.Dict().
			Int("hosts", stats.failedHosts).
			Int("metrics", stats.failedMetrics)).
		Msg("completed")
}

func (p *Poller) processedHostFromGanglia(grid Grid, cluster Cluster, host Host, now time.Time) (*domain.ProcessedHost, error) {
	dsm := domain.DSM{
		GridName:    grid.Name,
		ClusterName: cluster.Name,
		HostName:    host.Name,
	}
	hostId, ok := p.dsmRepo.GetHostId(dsm)
	if !ok {
		return nil, fmt.Errorf("hostId not known for %s", dsm)
	}
	pHost := domain.ProcessedHost{
		Id:      hostId,
		DSM:     dsm,
		Metrics: make(map[domain.MetricName]domain.ProcessedMetric),
	}
	return &pHost, nil
}

// MetricFromGanglia parses the given Ganglia metric into a format suitable
// for processing.
//
// The formats are largely similar with the following changes:
//
// Slope is replaced with a simplified Nature.
// TN is replaced with a more easily consumed Timestamp.
// TMAX is replaced with a more easily consumed Stale.
func processedMetricFromGanglia(now time.Time, src Metric) (domain.ProcessedMetric, error) {
	var dst domain.ProcessedMetric
	tn, err := strconv.Atoi(src.TN)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid TN")
	}
	tmax, err := strconv.Atoi(src.TMax)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid TMAX")
	}
	dmax, err := strconv.Atoi(src.DMax)
	if err != nil {
		return domain.ProcessedMetric{}, errors.Wrap(err, "invalid DMAX")
	}

	var stale bool
	persistent := dmax == 0
	if persistent {
		stale = tn > tmax*2
	} else {
		stale = tn > dmax
	}

	nature := "volatile"
	if src.Type == "string" || src.Type == "timestamp" {
		nature = "string_and_time"
	} else if src.Slope == "zero" {
		nature = "constant"
	}

	dst.Name = src.Name
	dst.Datatype = src.Type
	dst.Units = src.Units
	dst.Value = src.Val
	dst.Nature = nature
	dst.Dmax = dmax
	dst.Timestamp = now.Unix() - int64(tn)
	dst.Stale = stale

	return dst, nil
}
