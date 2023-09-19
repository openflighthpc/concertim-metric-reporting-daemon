// Package config holds the configuraiton for the application
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is the configuration struct for the app.
type Config struct {
	LogLevel         string `yaml:"log_level"`
	SharedSecretFile string `yaml:"shared_secret_file"`
	API              `yaml:"api"`
	DSM              `yaml:"dsm"`
	GDS              `yaml:"gds"`
	Retrieval        `yaml:"retrieval"`
	VisualizerAPI    `yaml:"visualizerAPI"`
}

// API is the configuration for the HTTP API component.
type API struct {
	IP        string        `yaml:"ip"`
	JWTSecret []byte        `yaml:"-"`
	Port      int           `yaml:"port"`
	Timeout   time.Duration `yaml:"timeout"`
}

// GDS is the configuration for the Ganglia Data Source server component.
type GDS struct {
	IP           string `yaml:"ip"`
	ClusterName  string `yaml:"clusterName"`
	Port         int    `yaml:"port"`
	MetricSource string `yaml:"metricSource"`
	HostTTL      int    `yaml:"hostTTL"`
}

// DSM is the configuration for the Data Source Map component.
type DSM struct {
	Duration  time.Duration `yaml:"duration"`
	Frequency time.Duration `yaml:"frequency"`
	Testdata  string        `yaml:"testdata"`
	Throttle  time.Duration `yaml:"throttle"`
}

// Retrieval is the configuration for retrieving the ganglia XML.
type Retrieval struct {
	ClusterName     string        `yaml:"clusterName"`
	Frequency       time.Duration `yaml:"frequency"`
	GridName        string        `yaml:"gridName"`
	IP              string        `yaml:"ip"`
	Port            int           `yaml:"port"`
	PostGmetadDelay time.Duration `yaml:"post_gmetad_delay"`
	Testdata        string        `yaml:"testdata"`
	Throttle        time.Duration `yaml:"throttle"`
}

type VisualizerAPI struct {
	AuthUrl              string `yaml:"authUrl"`
	DataSourceMapUrl     string `yaml:"data_source_map_url"`
	JWTSecret            []byte `yaml:"-"`
	Password             string `yaml:"password"`
	SkipCertificateCheck bool   `yaml:"skip_certificate_check"`
	Username             string `yaml:"username"`
}

// DefaultPath is the path to the default config file.
const DefaultPath string = "/opt/concertim/opt/ct-metric-reporting-daemon/config/config.yml"

// FromFile parses the given file path and returns a Config.
func FromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "loading config")
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("loading config from %s", path))
	}
	secret, err := ioutil.ReadFile(config.SharedSecretFile)
	if err != nil {
		return nil, errors.Wrap(err, "error reading shared secret")
	}
	config.API.JWTSecret = bytes.TrimRight(secret, "\n")
	config.VisualizerAPI.JWTSecret = bytes.TrimRight(secret, "\n")
	return &config, nil
}
