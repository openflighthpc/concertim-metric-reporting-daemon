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
	"github.com/alces-flight/concertim-metric-reporting-daemon/visualizer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Repo is an in-memory repository of device names to data source maps to
// host.
type Repo struct {
	config           config.DSM
	hostIdToDSM      map[domain.HostId]domain.DSM
	dsmToMemcacheKey map[domain.DSM]domain.MemcacheKey
	mux              sync.Mutex
	Ticker           *ticker.Ticker
	logger           zerolog.Logger
	visualizerClient *visualizer.Client
}

// DataRetriever is an interface for retrieving updated data for the Repo.
type dataRetriever interface {
	getNewData() (map[domain.HostId]domain.DSM, map[domain.DSM]domain.MemcacheKey, error)
}

// GetDSM returns the data source map for the given host name.
//
// See domain.DataSourceMapRepository interface for more details.
func (r *Repo) GetDSM(hostId domain.HostId) (domain.DSM, bool) {
	dsm, ok := r.getDSM(hostId)
	if !ok && r.Ticker.TickNow() {
		time.Sleep(r.config.Duration)
		dsm, ok = r.getDSM(hostId)
	}
	if !ok {
		r.logger.Debug().Stringer("lookup", hostId).Msg("not found")
	} else {
		r.logger.Debug().Stringer("lookup", hostId).Stringer("dsm", dsm).Msg("found")
	}
	return dsm, ok
}

func (r *Repo) getDSM(hostId domain.HostId) (domain.DSM, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	dsm, ok := r.hostIdToDSM[hostId]
	return dsm, ok
}

// Get looks up the given DSM and returns the stored memcache key for it.
//
// A second boolean value is returned indicating if the DSM was found, similar
// to indexing into a map.
func (r *Repo) GetMemcacheKey(dsm domain.DSM) (domain.MemcacheKey, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	memcacheKey, ok := r.dsmToMemcacheKey[dsm]
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
	newHostIdToDSM, newDSMToMemcacheKey, err := retriever.getNewData()
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	r.setData(newHostIdToDSM, newDSMToMemcacheKey)
	return nil
}

// New returns a new Repo.  It will be populated with assuming that the data
// retriever can do so.
func New(logger zerolog.Logger, config config.DSM, visualizerClient *visualizer.Client) *Repo {
	r := &Repo{
		config:           config,
		hostIdToDSM:      map[domain.HostId]domain.DSM{},
		dsmToMemcacheKey: map[domain.DSM]domain.MemcacheKey{},
		mux:              sync.Mutex{},
		Ticker:           ticker.NewTicker(config.Frequency, config.Throttle),
		logger:           logger.With().Str("component", "dsm-repo").Logger(),
		visualizerClient: visualizerClient,
	}
	r.runPeriodicUpdate()
	return r
}

func (r *Repo) setData(newHostIdToDSM map[domain.HostId]domain.DSM, newDSMToMemcacheKey map[domain.DSM]domain.MemcacheKey) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.hostIdToDSM = newHostIdToDSM
	r.dsmToMemcacheKey = newDSMToMemcacheKey
	r.logger.Info().
		Int("hostIdToDSM.count", len(newHostIdToDSM)).
		Int("dsmToMemcacheKey.count", len(newDSMToMemcacheKey)).
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
	case "visualizerAPI":
		return &visualizerAPIRetriever{
			client: r.visualizerClient,
			logger: r.logger,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown data retriever type: %s", r.config.Retriever)
	}
}
