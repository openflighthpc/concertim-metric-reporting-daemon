// Package dsmRepository contains an implementation of
// DataSourceMapRepository.  DataSourceMapRepository holds a map from device
// name to data source map to host.
//
// Device name is a user-friendly name used in the display of the appliance.
// Data source map to host may or may not be user-friendly; it is used only
// internally.  Examples of the mapping are below.
//
// * comp001      -> comp01.concertim.alces-flight.com
// * tempsensor01 -> sensor-dd5fb19b50624b33e1c5e4d5003714f4
// * pdu01        -> rack_2__powerstrip__startu42__1669827901
package dsmRepository

import (
	"fmt"
	"sync"
	"time"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/alces-flight/concertim-metric-reporting-daemon/ticker"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Repo is an in-memory repository of device names to data source maps to
// host.
type Repo struct {
	config      config.DSM
	hostnameMap map[domain.Hostname]domain.DSM
	memcacheMap map[domain.DSM]domain.MemcacheKey
	mux         sync.Mutex
	Ticker      *ticker.Ticker
	logger      zerolog.Logger
}

// DataRetriever is an interface for retrieving updated data for the Repo.
type dataRetriever interface {
	getNewData() (map[domain.Hostname]domain.DSM, map[domain.DSM]domain.MemcacheKey, error)
}

// GetDSM returns the data source map for the given host name.
//
// See domain.DataSourceMapRepository interface for more details.
func (r *Repo) GetDSM(hostname domain.Hostname) (domain.DSM, bool) {
	dsm, ok := r.getDSM(hostname)
	if !ok && r.Ticker.TickNow() {
		time.Sleep(r.config.Duration)
		dsm, ok = r.getDSM(hostname)
	}
	if !ok {
		r.logger.Debug().Stringer("lookup", hostname).Msg("not found")
	} else {
		r.logger.Debug().Stringer("lookup", hostname).Stringer("dsm", dsm).Msg("found")
	}
	return dsm, ok
}

func (r *Repo) getDSM(hostname domain.Hostname) (domain.DSM, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	dsm, ok := r.hostnameMap[hostname]
	return dsm, ok
}

// Get looks up the given DSM and returns the stored memcache key for it.
//
// A second boolean value is returned indicating if the DSM was found, similar
// to indexing into a map.
func (r *Repo) GetMemcacheKey(dsm domain.DSM) (domain.MemcacheKey, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	memcacheKey, ok := r.memcacheMap[dsm]
	if !ok {
		r.logger.Debug().Stringer("lookup", dsm).Msg("not found")
	} else {
		r.logger.Debug().Stringer("lookup", dsm).Stringer("memcacheKey", memcacheKey).Msg("found")
	}
	return memcacheKey, ok
}

// Update retrieves the latest DSM from an external source and updates its
// internal repository.
//
// The external source to use is configured when creating a new DSMRepo.
func (r *Repo) Update() error {
	retriever, err := r.getRetriver()
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	newHostMap, newMecacheMap, err := retriever.getNewData()
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	r.setData(newHostMap, newMecacheMap)
	return nil
}

// New returns a new Repo.  It will be populated with assuming that the data
// retriever can do so.
func New(logger zerolog.Logger, config config.DSM) *Repo {
	r := &Repo{
		config:      config,
		hostnameMap: map[domain.Hostname]domain.DSM{},
		memcacheMap: map[domain.DSM]domain.MemcacheKey{},
		mux:         sync.Mutex{},
		Ticker:      ticker.NewTicker(config.Frequency, config.Throttle),
		logger:      logger.With().Str("component", "dsm-repo").Logger(),
	}
	r.runPeriodicUpdate()
	return r
}

func (r *Repo) setData(newHostMap map[domain.Hostname]domain.DSM, newMecacheMap map[domain.DSM]domain.MemcacheKey) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.hostnameMap = newHostMap
	r.memcacheMap = newMecacheMap
	r.logger.Info().
		Int("hostmap.count", len(newHostMap)).
		Int("memcachemap.count", len(newMecacheMap)).
		Msg("updated")
}

func (r *Repo) runPeriodicUpdate() {
	go func() {
		r.logger.Debug().Dur("frequency", r.config.Frequency).Msg("Starting periodic retreival")
		for {
			<-r.Ticker.C
			err := r.Update()
			if err != nil {
				r.logger.Warn().Err(err).Msg("periodic update failed")
			}
		}
	}()
}

func (r *Repo) getRetriver() (dataRetriever, error) {
	switch r.config.Retriever {
	case "file":
		return &JSONFileRetreiver{
			Path:   r.config.Path,
			Logger: r.logger,
		}, nil
	case "script":
		return &Script{
			Args:   r.config.Args,
			Path:   r.config.Path,
			Logger: r.logger,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown data retriever type: %s", r.config.Retriever)
	}
}
