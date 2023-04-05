package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"

	"github.com/alces-flight/concertim-metric-reporting-daemon/processing/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// DSMRepo is an in-memory repository mapping from a ganglia host identifier
// (i.e., the triplet gridname/clustername/hostname) to a memcache device key
// (i.e., "hacor:<device class name>:<device id>").
//
// E.g., on my current dev machine this would hold a map from
// "unspecified"/"unspecified"/"comp203" to "hacor:device:9".
type DSMRepo struct {
	config config.DSM
	data   map[DSM]MemcacheKey
	logger zerolog.Logger
	mux    sync.Mutex
}

// NewDSMRepo returns a new empty *DSMRepo.  The Update method should be
// called before using it.
func NewDSMRepo(logger zerolog.Logger, config config.DSM) *DSMRepo {
	return &DSMRepo{
		config: config,
		logger: logger.With().Str("component", "dsm").Logger(),
	}
}

// Get looks up the given DSM and returns the stored memcache key for it.
//
// A second boolean value is returned indicating if the DSM was found, similar
// to indexing into a map.
func (r *DSMRepo) Get(dsm DSM) (MemcacheKey, bool) {
	memcacheKey, ok := r.data[dsm]
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
func (r *DSMRepo) Update() error {
	retriever, err := r.getRetriver()
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	newData, err := retriever.getNewData()
	if err != nil {
		return errors.Wrap(err, "updating DSM")
	}
	r.setData(newData)
	return nil
}

func (r *DSMRepo) setData(newData map[DSM]MemcacheKey) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.data = newData
	r.logger.Info().Int("count", len(newData)).Msg("set data")
}

func (r *DSMRepo) getRetriver() (dataRetriever, error) {
	switch r.config.Retriever {
	case "file":
		return &jsonFileRetreiver{
			path:   r.config.Path,
			logger: r.logger,
		}, nil
	case "script":
		return &scriptRetriever{
			path:   r.config.Path,
			logger: r.logger,
		}, nil
	default:
		return nil, fmt.Errorf("Unknown data retriever type: %s", r.config.Retriever)
	}
}

// dataRetriever is an interface for retrieving updated data for the DSMRepo.
type dataRetriever interface {
	getNewData() (map[DSM]MemcacheKey, error)
}

// jsonFileRetreiver retrieves the data source map from a pre-poulated JSON
// file.
type jsonFileRetreiver struct {
	path   string
	logger zerolog.Logger
}

func (j *jsonFileRetreiver) getNewData() (map[DSM]MemcacheKey, error) {
	j.logger.Debug().Str("path", j.path).Msg("retrieving")
	data, err := ioutil.ReadFile(j.path)
	if err != nil {
		msg := "reading JSON file"
		if !strings.Contains(err.Error(), j.path) {
			msg = fmt.Sprintf("%s: %s", msg, j.path)
		}
		return nil, errors.Wrap(err, msg)
	}
	return parseJSON(j.logger, data)
}

type scriptRetriever struct {
	path   string
	logger zerolog.Logger
}

func (sr *scriptRetriever) getNewData() (map[DSM]MemcacheKey, error) {
	sr.logger.Debug().Str("path", sr.path).Msg("retrieving")
	out, err := exec.Command(sr.path).Output()
	if err != nil {
		msg := "executing script"
		if !strings.Contains(err.Error(), sr.path) {
			msg = fmt.Sprintf("%s: %s", msg, sr.path)
		}
		return nil, errors.Wrap(err, msg)
	}
	return parseJSON(sr.logger, out)
}

// DSM represents a Ganglia identifier for a host.
type DSM struct {
	GridName    string
	ClusterName string
	HostName    string
}

func (d DSM) String() string {
	return fmt.Sprintf("%s/%s/%s", d.GridName, d.ClusterName, d.HostName)
}

// MemcacheKey exists to document some function signatures.
type MemcacheKey string

func (m MemcacheKey) String() string {
	return string(m)
}

func parseJSON(logger zerolog.Logger, data []byte) (map[DSM]MemcacheKey, error) {
	dsmMap := map[DSM]MemcacheKey{}
	gridMap := map[string]map[string]map[string]string{}
	err := json.Unmarshal(data, &gridMap)
	if err != nil {
		return nil, errors.Wrap(err, "parsing grid map")
	}
	for gName, clusterMap := range gridMap {
		for cName, hostMap := range clusterMap {
			for hName, key := range hostMap {
				dsm := DSM{gName, cName, hName}
				logger.Debug().
					Stringer("dsm", dsm).
					Str("memcacheKey", key).
					Msg("adding")
				dsmMap[dsm] = MemcacheKey(key)
			}
		}
	}
	return dsmMap, nil
}
