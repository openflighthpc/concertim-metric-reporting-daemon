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
	GDS              `yaml:"gds"`
	DSM              `yaml:"dsm"`
}

// API is the configuration for the HTTP API component.
type API struct {
	IP        string `yaml:"ip"`
	Port      int    `yaml:"port"`
	JWTSecret []byte `yaml:"-"`
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
	Retriever string        `yaml:"retriever"`
	Path      string        `yaml:"path"`
	Args      []string      `yaml:"args"`
	Sleep     time.Duration `yaml:"sleep"`
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
	return &config, nil
}
