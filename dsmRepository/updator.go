package dsmRepository

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/alces-flight/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Updater struct {
	config    config.DSM
	limiter   rate.Sometimes
	logger    zerolog.Logger
	repo      domain.DataSourceMapRepository
	retriever domain.DataSourceMapRetreiver
	ticker    *time.Ticker
}

// New returns a new Updater.  It will be populated with assuming that the data
// retriever can do so.
func NewUpdater(
	logger zerolog.Logger,
	config config.DSM,
	repo domain.DataSourceMapRepository,
	retriever domain.DataSourceMapRetreiver,
) *Updater {
	logger = logger.With().Str("component", "dsm-updater").Logger()
	u := &Updater{
		config:    config,
		repo:      repo,
		ticker:    time.NewTicker(config.Frequency),
		limiter:   rate.Sometimes{First: 1, Interval: config.Throttle},
		logger:    logger,
		retriever: retriever,
	}
	u.RunPeriodicUpdateLoop()
	return u
}

func (u *Updater) RunPeriodicUpdateLoop() {
	go func() {
		u.logger.Debug().Dur("frequency", u.config.Frequency).Msg("Starting periodic retreival")
		for {
			<-u.ticker.C
			u.limiter.Do(func() {
				err := u.update()
				if err != nil {
					u.logger.Warn().Err(err).Msg("periodic update failed")
				}
			})
		}
	}()
}

func (u *Updater) UpdateNow() {
	u.limiter.Do(func() {
		err := u.update()
		if err != nil {
			u.logger.Warn().Err(err).Msg("on demand update failed")
		}
	})
}

// Update retrieves the latest DSM from an external source and updates its
// internal repository.
//
// The external source to use is configured when creating a new DSMRepo.
func (u *Updater) update() error {
	hostIdToDSM, dsmToHostId, err := u.retriever.GetDSM()
	if err != nil {
		return errors.Wrap(err, "retrieving DSM")
	}
	err = u.repo.Update(hostIdToDSM, dsmToHostId)
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	return nil
}
