package dsmRepository

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Repo is an in-memory repository of device names to data source maps to
// host.
type Repo struct {
	data   map[string]string
	mux    sync.Mutex
	logger zerolog.Logger
}

// DataRetriever is an interface for retrieving updated data for the Repo.
type dataRetriever interface {
	getNewData() (map[string]string, error)
}

// Get returns the data source map to host for the given device name.
//
// deviceName is a user-friendly name used in the display of the
// appliance.  The returned mapToHost may or may not be user-friendly it
// is used only internally.  Examples of the mapping are below.
//
// * comp001      -> comp01.concertim.alces-flight.com
// * tempsensor01 -> sensor-dd5fb19b50624b33e1c5e4d5003714f4
// * pdu01        -> rack_2__powerstrip__startu42__1669827901
func (r *Repo) Get(deviceName string) (string, bool) {
	r.mux.Lock()
	defer r.mux.Unlock()
	mapToHost, ok := r.data[deviceName]
	r.logger.Debug().Str("lookup", deviceName).Str("found", mapToHost).Send()
	return mapToHost, ok
}

// New returns a new Repo.  It will be populated with assuming that the data
// retriever can do so.
func New(logger zerolog.Logger) *Repo {
	r := &Repo{
		data:   map[string]string{},
		mux:    sync.Mutex{},
		logger: logger.With().Str("component", "dsm-repo").Logger(),
	}
	retriever := JSONFileRetreiver{
		Path:   "/home/ben/projects/concertim-deploy/ansible/tmp/foo.json",
		Logger: logger.With().Str("compent", "dsm-repo data retriever").Logger(),
	}
	r.runUpdateTimer(&retriever)
	return r
}

func (r *Repo) setData(newData map[string]string) {
	r.logger.Debug().Msg("Updating data")
	r.mux.Lock()
	defer r.mux.Unlock()
	r.data = newData
}

func (r *Repo) runUpdateTimer(retriever dataRetriever) {
	go func() {
		for {
			newData, err := retriever.getNewData()
			if err != nil {
				r.logger.Warn().Err(err).Msg("Unable to update")
				continue
			}
			r.setData(newData)
			time.Sleep(60 * time.Second)
		}
	}()
}
