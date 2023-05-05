// Package config holds the configuraiton for the application
package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"

	topConfig "github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/pkg/errors"
)

// Config is the configuration struct for the app.
type Config struct {
	LogLevel  string        `yaml:"log_level"`
	DSM       topConfig.DSM `yaml:"dsm"`
	Retrieval `yaml:"retrieval"`
	Recorder  `yaml:"recorder"`
}

// Retrieval is the configuration for retrieving the ganglia XML.
type Retrieval struct {
	IP       string        `yaml:"ip"`
	Port     int           `yaml:"port"`
	Sleep    time.Duration `yaml:"sleep"`
	Testdata string        `yaml:"testdata"`
}

// Recorder is the configuration for recording the processed results.
type Recorder struct {
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

// DefaultPath is the path to the default config file.
const DefaultPath string = "/opt/concertim/opt/ct-metric-processing-daemon/config/config.yml"

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
	return &config, nil
}
