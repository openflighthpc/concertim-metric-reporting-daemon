// Package config holds the configuraiton for the application
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config is the configuration struct for the app.
type Config struct {
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
}

// DSM is the configuration for the Data Source Map component.
type DSM struct {
	Retriever string `yaml:"retriever"`
	Path      string `yaml:"path"`
	Sleep     int64  `yaml:"sleep"`
}

// FromFile parses the given file path and returns a Config.
func FromFile(path string) (*Config, error) {
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
