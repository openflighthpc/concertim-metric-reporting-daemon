// Package config holds the configuraiton for the application
package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the configuration struct for the app.
type Config struct {
	LogLevel string `yaml:"log_level"`
	API `yaml:"api"`
	GDS `yaml:"gds"`
	DSM `yaml:"dsm"`
}

// API is the configuration for the HTTP API component.
type API struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
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
	Retriever string `yaml:"retriever"`
	Path      string `yaml:"path"`
	Sleep     int    `yaml:"sleep"`
}

// DefaultPaths contains the default paths used to search for a config file.
var DefaultPaths = []string{
	"/data/private/share/daemons/ct-metric-reporting-daemon/config.yml",
	"./config.yml",
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
