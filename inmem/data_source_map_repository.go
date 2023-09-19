package inmem

import (
	"sync"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/rs/zerolog"
)

// DSMRepo is an in-memory repository of device IDs to data source maps to host
// and vice-versa.
type DSMRepo struct {
	config      config.DSM
	hostIdToDSM map[domain.HostId]domain.DSM
	dsmToHostId map[domain.DSM]domain.HostId
	mux         sync.Mutex
	logger      zerolog.Logger
}

var _ domain.DataSourceMapRepository = (*DSMRepo)(nil)

// NewDSMRepo returns a new empty DSMRepo.
func NewDSMRepo(logger zerolog.Logger, config config.DSM) *DSMRepo {
	return &DSMRepo{
		config:      config,
		hostIdToDSM: map[domain.HostId]domain.DSM{},
		dsmToHostId: map[domain.DSM]domain.HostId{},
		mux:         sync.Mutex{},
		logger:      logger.With().Str("component", "dsm-repo").Logger(),
	}
}

// GetDSM returns the data source map for the given host name.
//
// See domain.DataSourceMapRepository interface for more details.
func (r *DSMRepo) GetDSM(hostId domain.HostId) (domain.DSM, bool) {
	dsm, ok := r.getDSM(hostId)
	// if !ok && r.Ticker.TickNow() {
	// 	time.Sleep(r.config.Duration)
	// 	dsm, ok = r.getDSM(hostId)
	// }
	if !ok {
		r.logger.Debug().Stringer("lookup", hostId).Msg("not found")
	} else {
		r.logger.Debug().Stringer("lookup", hostId).Stringer("dsm", dsm).Msg("found")
	}
	return dsm, ok
}

func (r *DSMRepo) getDSM(hostId domain.HostId) (domain.DSM, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	dsm, ok := r.hostIdToDSM[hostId]
	return dsm, ok
}

// GetHostId looks up the given DSM and returns the stored host id.
//
// A second boolean value is returned indicating if the DSM was found, similar
// to indexing into a map.
func (r *DSMRepo) GetHostId(dsm domain.DSM) (domain.HostId, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	hostId, ok := r.dsmToHostId[dsm]
	if !ok {
		r.logger.Debug().Stringer("lookup", dsm).Msg("not found")
	} else {
		r.logger.Debug().Stringer("lookup", dsm).Stringer("hostId", hostId).Msg("found")
	}
	return hostId, ok
}

// Update the state of the repository with the given data.
func (r *DSMRepo) Update(newHostIdToDSM map[domain.HostId]domain.DSM, newDSMToHostId map[domain.DSM]domain.HostId) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.hostIdToDSM = newHostIdToDSM
	r.dsmToHostId = newDSMToHostId
	r.logger.Info().
		Int("hostIdToDSM.count", len(newHostIdToDSM)).
		Int("dsmToHostId.count", len(newDSMToHostId)).
		Msg("updated")
	return nil
}
