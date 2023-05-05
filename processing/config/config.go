// Package config holds the configuraiton for the application
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	topConfig "github.com/alces-flight/concertim-metric-reporting-daemon/config"
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

// DefaultPaths contains the default paths used to search for a config file.
var DefaultPaths = []string{
	"./config/config.yml",
	"/opt/concertim/opt/ct-metric-processing-daemon/config/config.yml",
}

// FromFile parses the given file path and returns a Config.
func FromFile(paths []string) (*Config, error) {
	path, err := findConfigFile(paths)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func findConfigFile(paths []string) (string, error) {
	for _, path := range paths {
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("Unable to find config file")
}
